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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func NewInstance(id, workingDirectory string) (string, error) {
	cfg := GetConfig()

	if id == "" {
		id = model.Id

		_, found := cfg.Instances[id]
		idx := 2
		for found {
			id = fmt.Sprintf("%v-%v", model.Id, idx)
			idx++
			_, found = cfg.Instances[id]
		}
	}

	if _, found := cfg.Instances[id]; found {
		return "", errors.Errorf("instance with id %v already exists", id)
	}

	if workingDirectory == "" {
		root := userInstanceRoot()
		workingDirectory = filepath.Join(root, id)
	}

	if err := os.MkdirAll(workingDirectory, os.ModePerm); err != nil {
		return "", errors.Wrapf(err, "unable to create instance directory [%v]", workingDirectory)
	}

	instanceConfig := &InstanceConfig{
		Id:               id,
		Model:            model.Id,
		WorkingDirectory: workingDirectory,
	}

	cfg.Instances[id] = instanceConfig
	cfg.Default = id
	if err := PersistConfig(cfg); err != nil {
		return "", err
	}
	return id, nil
}

func SetActiveInstance(newInstanceId string) error {
	cfg := GetConfig()
	newInstanceConfig, found := cfg.Instances[newInstanceId]
	if !found {
		return errors.Errorf("invalid instance id [%s]", newInstanceId)
	}
	instanceConfig = newInstanceConfig

	if _, err := instanceConfig.LoadLabel(); err != nil {
		return errors.Wrapf(err, "invalid instance working directory [%v] for instance [%v]", instanceConfig.WorkingDirectory, instanceConfig.Id)
	}

	cfg.Default = newInstanceId
	if err := PersistConfig(cfg); err != nil {
		return errors.Wrapf(err, "unable to update active instance to [%s] in config file [%v]", newInstanceConfig, cfg.ConfigPath)
	}
	return nil
}

func ActiveInstancePath() string {
	GetActiveInstanceConfig()
	if instanceConfig != nil {
		return instanceConfig.WorkingDirectory
	}
	return ""
}

func ActiveInstanceId() string {
	GetConfig()
	return config.GetSelectedInstanceId()
}

func BootstrapInstance() error {
	logrus.Debugf("bootstrapping instance %v", ActiveInstanceId())

	if _, err := os.Stat(ActiveInstancePath()); err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("invalid active instance")
		} else {
			return err
		}
	}
	return nil
}

func instancePath(instanceId string) string {
	return filepath.Join(userInstanceRoot(), instanceId)
}

func userInstanceRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("unable to get user home directory (%v)", err)
	}
	return filepath.Join(home, ".fablab/instances")
}
