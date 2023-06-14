/*
	Copyright 2019 NetFoundry Inc.

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

package terraform_0

import (
	"fmt"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/resources"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func Express() model.Stage {
	return &terraform{}
}

type terraform struct{}

func (t *terraform) Execute(run model.Run) error {
	m := run.GetModel()
	l := run.GetLabel()

	if err := t.generate(m); err != nil {
		return err
	}
	if err := t.init(); err != nil {
		return err
	}
	if err := t.apply(); err != nil {
		return err
	}
	if err := t.bind(m, l); err != nil {
		return err
	}
	return nil
}

func (t *terraform) generate(m *model.Model) error {
	terraformResource := m.GetResource(resources.Terraform)

	visitor := &terraformVisitor{
		model:    m,
		resource: terraformResource,
	}

	if err := fs.WalkDir(terraformResource, ".", visitor.visit); err != nil {
		return errors.Wrapf(err, "error generating terraform")
	}
	return nil
}

func (t *terraform) init() error {
	prc := lib.NewProcess("terraform", "init")
	prc.Cmd.Dir = terraformRun()
	prc.WithTail(lib.StdoutTail)
	if err := prc.Run(); err != nil {
		return fmt.Errorf("error running 'terraform init' (%w)", err)
	}
	return nil
}

func (t *terraform) apply() error {
	prc := lib.NewProcess("terraform", "apply", "-auto-approve")
	prc.Cmd.Dir = terraformRun()
	prc.WithTail(lib.StdoutTail)
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
				return fmt.Errorf("unable to get output [%s] (%s)", publicIpOutput, err)
			}

			privateIpOutput := fmt.Sprintf("%s_host_%s_private_ip", regionId, hostId)
			if output, err := terraformOutput(privateIpOutput); err == nil {
				l.Bindings[privateIpOutput] = output
				logrus.Infof("set private ip [%s] for [%s/%s]", output, regionId, hostId)
			} else {
				return fmt.Errorf("unable to get output [%s] (%s)", privateIpOutput, err)
			}
		}
	}
	if err := l.Save(); err != nil {
		return fmt.Errorf("unable to save updated instance label [%s] (%w)", model.BuildPath(), err)
	}
	m.BindLabel(l)
	return nil
}

func (t *terraformVisitor) visit(path string, e fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	fi, err := e.Info()
	if err != nil {
		return err
	}

	if fi.Mode().IsRegular() {
		logrus.Debugf("visiting [%s]", path)

		outputPath := filepath.Join(terraformRun(), path)
		if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
			return fmt.Errorf("error creating parent directories [%s] (%w)", outputPath, err)
		}

		err = lib.RenderTemplateFS(t.resource, path, outputPath, t.model, struct {
			Model        *model.Model
			TerraformLib string
		}{
			Model:        t.model,
			TerraformLib: terraformRun(),
		})
		if err != nil {
			return errors.Wrap(err, "error rendering template")
		}

		logrus.Infof("=> [%s]", path)
	}
	return nil
}

type terraformVisitor struct {
	model    *model.Model
	resource fs.FS
}

func terraformOutput(name string) (string, error) {
	prc := lib.NewProcess("terraform", "output", name)
	prc.Cmd.Dir = terraformRun()
	if err := prc.Run(); err != nil {
		return "", fmt.Errorf("error executing 'terraform output' (%w)", err)
	}
	return strings.Trim(prc.Output.String(), " \t\r\n\""), nil
}

func terraformRun() string {
	return filepath.Join(model.BuildPath(), "tf")
}
