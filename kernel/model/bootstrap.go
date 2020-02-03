/*
	Copyright 2019 NetFoundry, Inc.

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
	"reflect"
)

func AddBootstrapExtension(ext BootstrapExtension) {
	bootstrapExtensions = append(bootstrapExtensions, ext)
}

func Bootstrap() error {
	var err error
	var m *Model
	if err = BootstrapBindings(); err != nil {
		return fmt.Errorf("unable to bootstrap config (%w)", err)
	}
	if err = BootstrapInstance(); err != nil {
		return fmt.Errorf("unable to bootstrap active instance (%w)", err)
	}
	if instanceId != "" {
		for _, ext := range bootstrapExtensions {
			if err := ext.Bootstrap(m); err != nil {
				return fmt.Errorf("unable to bootstrap extension (%w)", err)
			}
		}
		if err = bootstrapPaths(); err != nil {
			return fmt.Errorf("unable to bootstrap paths (%w)", err)
		}
		if err = bootstrapLabel(); err != nil {
			return fmt.Errorf("unable to bootstrap label (%w)", err)
		}
		if m, err = bootstrapModel(); err != nil {
			return fmt.Errorf("unable to bootstrap binding (%w)", err)
		}
	} else {
		logrus.Warnf("no active instance")
	}
	return nil
}

func BootstrapBindings() error {
	var err error
	if err = loadBindings(); err != nil {
		return fmt.Errorf("unable to bootstrap config (%w)", err)
	}
	return nil
}

type BootstrapExtension interface {
	Bootstrap(m *Model) error
}

func bootstrapModel() (*Model, error) {
	l := GetLabel()
	if l != nil {
		m, found := modelRegistry[l.Model]
		if !found {
			return nil, fmt.Errorf("no such model [%s]", l.Model)
		}

		if m.Parent != nil {
			if err := m.Merge(m.Parent); err != nil {
				return nil, fmt.Errorf("error merging parent (%w)", err)
			}
		}

		m.BindLabel(l)
		for _, factory := range m.Factories {
			if err := factory.Build(m); err != nil {
				return nil, fmt.Errorf("error executing factory [%s] (%w)", reflect.TypeOf(factory), err)
			}
		}
		m.BindBindings(bindings)

		m.infrastructureStages = nil
		for _, binder := range m.Infrastructure {
			m.infrastructureStages = append(m.infrastructureStages, binder(m))
		}
		m.configurationStages = nil
		for _, binder := range m.Configuration {
			m.configurationStages = append(m.configurationStages, binder(m))
		}
		m.kittingStages = nil
		for _, binder := range m.Kitting {
			m.kittingStages = append(m.kittingStages, binder(m))
		}
		m.distributionStages = nil
		for _, binder := range m.Distribution {
			m.distributionStages = append(m.distributionStages, binder(m))
		}
		m.activationStages = nil
		for _, binder := range m.Activation {
			m.activationStages = append(m.activationStages, binder(m))
		}
		m.operationStages = nil
		for _, binder := range m.Operation {
			m.operationStages = append(m.operationStages, binder(m))
		}
		m.disposalStages = nil
		for _, binder := range m.Disposal {
			m.disposalStages = append(m.disposalStages, binder(m))
		}

		m.actions = make(map[string]Action)
		for name, binder := range m.Actions {
			m.actions[name] = binder(m)
			logrus.Debugf("bound action [%s]", name)
		}

		return m, nil

	} else {
		logrus.Warn("no run label found")
	}
	return nil, nil
}
