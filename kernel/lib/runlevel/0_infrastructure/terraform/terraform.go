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
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/lib"
	semaphore_0 "github.com/openziti/fablab/kernel/lib/runlevel/0_infrastructure/semaphore"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/resources"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Express() model.Stage {
	return &Terraform{}
}

type Terraform struct {
	Retries    uint8
	ReadyCheck *semaphore_0.ReadyStage
}

func (t *Terraform) Execute(run model.Run) error {
	m := run.GetModel()
	l := run.GetLabel()

	if err := t.generate(m); err != nil {
		return err
	}

	attemptsRemaining := t.Retries + 1

	initDone := false

	var err error
	for attemptsRemaining > 0 {
		if !initDone {
			err = t.Init()
			if err == nil {
				initDone = true
			}
		}

		if err == nil {
			err = t.apply()
		}

		if err == nil {
			err = t.bind(m, l)
		}

		if err == nil && t.ReadyCheck != nil {
			err = t.ReadyCheck.Execute(run)
		}

		if err == nil {
			return nil
		}

		attemptsRemaining--
		if attemptsRemaining > 0 {
			pfxlog.Logger().WithError(err).Error("terraform failure, retrying in 3s")
			time.Sleep(3 * time.Second)
		}
	}

	return err
}

func (t *Terraform) generate(m *model.Model) error {
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

func (t *Terraform) Init() error {
	prc := lib.NewProcess("terraform", "init")
	prc.Cmd.Dir = terraformRun()
	prc.WithTail(lib.StdoutTail)
	if err := prc.Run(); err != nil {
		return fmt.Errorf("error running 'terraform init' (%w)", err)
	}
	return nil
}

func (t *Terraform) apply() error {
	prc := lib.NewProcess("terraform", "apply", "-auto-approve")
	prc.Cmd.Dir = terraformRun()
	prc.WithTail(lib.StdoutTail)
	if err := prc.Run(); err != nil {
		return fmt.Errorf("error running 'terraform apply' (%w)", err)
	}
	return nil
}

func (t *Terraform) bind(m *model.Model, l *model.Label) error {
	hostIps := map[string]string{}

	output, err := allTerraformOutput()
	if err != nil {
		return err
	}

	for regionId, region := range m.Regions {
		for hostId := range region.Hosts {
			publicIpKey := fmt.Sprintf("%s_host_%s_public_ip", regionId, hostId)
			publicIpVal, found := output[publicIpKey]
			if !found {
				return fmt.Errorf("unable to get public key for [%s]", publicIpKey)
			}
			l.Bindings[publicIpKey] = publicIpVal

			if otherHostId, found := hostIps[publicIpKey]; found {
				return errors.Errorf("duplicate ips found, terraform bug! ip %s found for hosts %s and %s",
					publicIpKey, otherHostId, hostId)
			}

			hostIps[publicIpKey] = hostId

			privateIpKey := fmt.Sprintf("%s_host_%s_private_ip", regionId, hostId)
			privateIpVal, found := output[privateIpKey]
			if !found {
				return fmt.Errorf("unable to get private key for [%s]", privateIpKey)
			}
			l.Bindings[privateIpKey] = privateIpVal
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

func allTerraformOutput() (map[string]string, error) {
	prc := lib.NewProcess("terraform", "output")
	prc.Cmd.Dir = terraformRun()
	if err := prc.Run(); err != nil {
		return nil, errors.Wrap(err, "error executing 'terraform output'")
	}
	result := map[string]string{}
	lines := strings.Split(prc.Output.String(), "\n")
	for _, line := range lines {
		line = strings.Trim(line, " \t\r\n\"")
		if line == "" {
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			return nil, errors.Errorf("error parsing 'terraform output' line '%s'", line)
		}
		key := strings.Trim(parts[0], " \t\r\n\"")
		val := strings.Trim(parts[1], " \t\r\n\"")
		result[key] = val
	}
	return result, nil
}

func terraformRun() string {
	return filepath.Join(model.BuildPath(), "tf")
}
