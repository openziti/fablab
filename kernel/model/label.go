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

package model

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

func (m *Model) BindLabel(l *Label) {
	clean := true
	for regionId, region := range m.Regions {
		for hostId, host := range region.Hosts {
			publicIpBinding := fmt.Sprintf("%s_host_%s_public_ip", regionId, hostId)
			if binding, found := l.Bindings[publicIpBinding]; found {
				if publicIp, ok := binding.(string); ok {
					host.PublicIp = publicIp
				}
			} else {
				logrus.Warnf("no binding [%s]", publicIpBinding)
				clean = false
			}

			privateIpBinding := fmt.Sprintf("%s_host_%s_private_ip", regionId, hostId)
			if binding, found := l.Bindings[privateIpBinding]; found {
				if privateIp, ok := binding.(string); ok {
					host.PrivateIp = privateIp
				}
			} else {
				logrus.Warnf("no binding [%s]", privateIpBinding)
				clean = false
			}
		}
	}
	if clean {
		m.bound = true
	}
}

func GetLabel() *Label {
	return label
}

func (label *Label) Save() error {
	return label.SaveAtPath(label.path)
}

func (label *Label) SaveAtPath(path string) error {
	data, err := yaml.Marshal(label)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	labelDir := filepath.Dir(labelPath(path))
	if err := os.MkdirAll(labelDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create label directory [%s] (%s)", labelDir, err)
	}

	if err = os.WriteFile(labelPath(path), data, 0600); err != nil {
		return err
	}

	return nil
}

func (label *Label) GetFilePath(fileName string) string {
	return filepath.Join(label.path, fileName)
}

func CreateLabel(instanceId string, bindings map[string]string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	workingDir := cfg.Instances[instanceId].WorkingDirectory

	if old, err := LoadLabel(workingDir); err == nil {
		return fmt.Errorf("existing instance [%s] found at [%s]", old.Model, workingDir)
	}

	l := &Label{
		InstanceId: instanceId,
		Model:      GetModel().GetId(),
		State:      Created,
		Bindings:   map[string]interface{}{},
	}

	for k, v := range bindings {
		l.Bindings[k] = v
	}

	if err = l.SaveAtPath(workingDir); err != nil {
		return fmt.Errorf("error writing run label [%s] (%s)", workingDir, err)
	}
	return nil
}

func LoadLabel(path string) (*Label, error) {
	data, err := os.ReadFile(filepath.Join(path, labelFilename))
	if err != nil {
		return nil, err
	}
	l := &Label{Bindings: Variables{}}
	if err = yaml.Unmarshal(data, l); err != nil {
		return nil, err
	}
	l.path = path
	return l, nil
}

func bootstrapLabel() error {
	instancePath := ActiveInstancePath()
	if _, err := os.Stat(labelPath(instancePath)); err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("no label at instance path [%s]", instancePath)
			return nil
		}
		return fmt.Errorf("unable to stat run label [%s] (%s)", labelPath(instancePath), err)
	}
	if l, err := LoadLabel(instancePath); err == nil {
		label = l
	} else {
		return fmt.Errorf("unable to bootstrap instance label [%s] (%s)", instancePath, err)
	}
	return nil
}

func labelPath(path string) string {
	return filepath.Join(path, labelFilename)
}

type Label struct {
	InstanceId string        `yaml:"id"`
	Model      string        `yaml:"model"`
	State      InstanceState `yaml:"state"`
	Bindings   Variables     `yaml:"bindings"`
	path       string
}

type InstanceState int

const (
	Created InstanceState = iota
	Expressed
	Configured
	Distributed
	Activated
	Operating
	Disposed
)

func (instanceState InstanceState) String() string {
	names := [...]string{
		"Created",
		"Expressed",
		"Configured",
		"Distributed",
		"Activated",
		"Operating",
		"Disposed",
	}
	if instanceState < Created || instanceState > Disposed {
		return "<<Invalid>>"
	}
	return names[instanceState]
}

const labelFilename = "fablab.yml"
