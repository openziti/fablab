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

package model

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func NewNamedInstance(name string) error {
	root := userInstanceRoot()
	dir := filepath.Join(root, name)

	if err := createUserInstanceRoot(); err != nil {
		return fmt.Errorf("unable to create instance root [%s] (%w)", dir, err)
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create instance root [%s] (%w)", dir, err)
	}

	return nil
}

func NewInstance() (string, error) {
	if err := createUserInstanceRoot(); err != nil {
		return "", fmt.Errorf("unable to create instance root (%w)", err)
	}

	root := userInstanceRoot()
	dir, err := ioutil.TempDir(root, "")
	if err != nil {
		return "", fmt.Errorf("unable to allocate directory [%s] (%w)", root, err)
	}
	return filepath.Base(dir), nil
}

func ListInstances() ([]string, error) {
	root := userInstanceRoot()

	instances := make([]string, 0)
	if _, err := os.Stat(root); err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("unable to stat instance root [%s] (%w)", root, err)
		}
	}
	ids, err := ioutil.ReadDir(root)
	if err == nil {
		for _, id := range ids {
			if id.IsDir() {
				instances = append(instances, id.Name())
			}
		}
	} else {
		logrus.Warnf("no instance root [%s]", root)
	}
	return instances, nil
}

func RemoveInstance(instanceId string) error {
	path := instancePath(instanceId)
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("error remove instance [%s] (%w)", instanceId, err)
	}
	return nil
}

func SetActiveInstance(instanceId string) error {
	if _, err := LoadLabelForInstance(instanceId); err != nil {
		return fmt.Errorf("invalid instance path [%s] (%w)", instancePath(instanceId), err)
	}
	if err := ioutil.WriteFile(activeInstance(), []byte(instanceId), os.ModePerm); err != nil {
		return fmt.Errorf("unable to store active instance [%s] (%w)", activeInstance(), err)
	}
	instanceId = instanceId
	return nil
}

func ClearActiveInstance() error {
	if err := ioutil.WriteFile(activeInstance(), []byte(""), os.ModePerm); err != nil {
		return fmt.Errorf("unable to clear active instance [%s] (%w)", activeInstance(), err)
	}
	instanceId = ""
	return nil
}

func ActiveInstancePath() string {
	return filepath.Join(userInstanceRoot(), instanceId)
}

func ActiveInstanceId() string {
	return instanceId
}

func BootstrapInstance() error {
	var err error
	if instanceId, err = loadActiveInstance(); err != nil {
		return fmt.Errorf("unable to load active instance (%w)", err)
	}
	if _, err := os.Stat(ActiveInstancePath()); err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("invalid active instance")
		} else {
			return err
		}
	}
	return nil
}

func loadActiveInstance() (string, error) {
	var data []byte
	var err error
	data, err = ioutil.ReadFile(activeInstance())
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("no active instance [%s]", activeInstance())
		} else {
			return "", fmt.Errorf("error reading active instance [%s] (%w)", activeInstance(), err)
		}
	}

	path := strings.Trim(string(data), " \t\r\n")
	return path, nil
}

func instancePath(instanceId string) string {
	return filepath.Join(userInstanceRoot(), instanceId)
}

func activeInstance() string {
	return filepath.Join(configRoot(), "active-instance")
}

func createUserInstanceRoot() error {
	root := userInstanceRoot()
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(root, os.ModePerm); err != nil {
				return fmt.Errorf("unable to create instance root [%s] (%w)", root, err)
			}
		} else {
			return fmt.Errorf("unable to stat instance root [%s] (%w)", root, err)
		}
	}
	return nil
}

func userInstanceRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("unable to get user home directory (%w)", err)
	}
	return filepath.Join(home, ".fablab/instances")
}
