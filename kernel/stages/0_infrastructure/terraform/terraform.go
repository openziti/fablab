/*
	Copyright 2019 Netfoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package terraform

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel"
	"github.com/netfoundry/fablab/model"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func Express() model.InfrastructureStage {
	return &terraform{}
}

func (t *terraform) Express(m *model.Model, l *model.Label) error {
	if err := t.generate(m); err != nil {
		return fmt.Errorf("%w", err)
	}
	if err := t.init(); err != nil {
		return fmt.Errorf("%w", err)
	}
	if err := t.apply(); err != nil {
		return fmt.Errorf("%w", err)
	}
	if err := t.bind(m, l); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (t *terraform) generate(m *model.Model) error {
	visitor := &terraformVisitor{model: m}
	if err := filepath.Walk(terraformSrc(), visitor.visit); err != nil {
		return fmt.Errorf("error generating terraform (%w)", err)
	}
	return nil
}

func (t *terraform) init() error {
	prc := kernel.NewProcess("terraform", "init")
	prc.Cmd.Dir = terraformRun()
	prc.WithTail(kernel.StdoutTail)
	if err := prc.Run(); err != nil {
		return fmt.Errorf("error running 'terraform init' (%w)", err)
	}
	return nil
}

func (t *terraform) apply() error {
	prc := kernel.NewProcess("terraform", "apply", "-auto-approve")
	prc.Cmd.Dir = terraformRun()
	prc.WithTail(kernel.StdoutTail)
	if err := prc.Run(); err != nil {
		return fmt.Errorf("error running 'terraform apply' (%w)", err)
	}
	return nil
}

func (t *terraform) bind(m *model.Model, l *model.Label) error {
	for regionId, region := range m.Regions {
		for hostId := range region.Hosts {
			publicIpOutput := fmt.Sprintf("%s_host_%s_public_ip", regionId, hostId)
			if output, err := terraformOutput(publicIpOutput); err == nil {
				l.Bindings[publicIpOutput] = output
				logrus.Infof("set public ip [%s] for [%s/%s]", output, regionId, hostId)
			} else {
				logrus.Fatalf("unable to get output [%s] (%s)", publicIpOutput, err)
			}

			privateIpOutput := fmt.Sprintf("%s_host_%s_private_ip", regionId, hostId)
			if output, err := terraformOutput(privateIpOutput); err == nil {
				l.Bindings[privateIpOutput] = output
				logrus.Infof("set private ip [%s] for [%s/%s]", output, regionId, hostId)
			} else {
				logrus.Fatalf("unable to get output [%s] (%s)", privateIpOutput, err)
			}
		}
	}
	if err := l.Save(); err != nil {
		logrus.Fatalf("unable to save updated instance label [%s] (%w)", model.ActiveInstancePath(), err)
	}
	m.BindLabel(l)
	return nil
}

type terraform struct {
}

func (t *terraformVisitor) visit(path string, fi os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if fi.Mode().IsRegular() {
		logrus.Debugf("visiting [%s]", path)

		rel, err := filepath.Rel(terraformSrc(), path)
		if err != nil {
			return fmt.Errorf("error relativizing path [%s] (%w)", path, err)
		}

		tp, err := template.ParseFiles(path)
		if err != nil {
			return fmt.Errorf("error parsing template [%s] (%w)", path, err)
		}

		outputPath := filepath.Join(terraformRun(), rel)
		if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
			return fmt.Errorf("error creating parent directories [%s] (%w)", outputPath, err)
		}

		outputF, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating terraform output [%s] (%w)", outputPath, err)
		}
		defer func() { _ = outputF.Close() }()

		err = tp.Execute(outputF, struct {
			Model        *model.Model
			TerraformLib string
		}{
			Model:        t.model,
			TerraformLib: terraformLib(),
		})
		if err != nil {
			return err
		}

		logrus.Infof("=> [%s]", rel)
	}
	return nil
}

type terraformVisitor struct {
	model *model.Model
}

func terraformOutput(name string) (string, error) {
	prc := kernel.NewProcess("terraform", "output", name)
	prc.Cmd.Dir = terraformRun()
	if err := prc.Run(); err != nil {
		return "", fmt.Errorf("error executing 'terraform output' (%w)", err)
	}
	return strings.Trim(prc.Output.String(), " \t\r\n"), nil
}

func terraformSrc() string {
	return filepath.Join(model.FablabRoot(), "lib/templates/tf")
}

func terraformLib() string {
	return filepath.Join(model.FablabRoot(), "lib/tf")
}

func terraformRun() string {
	return filepath.Join(model.ActiveInstancePath(), "tf")
}
