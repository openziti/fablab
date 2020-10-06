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
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "unable to bootstrap config")
	}
	if err = BootstrapInstance(); err != nil {
		return errors.Wrap(err, "unable to bootstrap active instance")
	}
	if instanceId != "" {
		for _, ext := range bootstrapExtensions {
			if err := ext.Bootstrap(m); err != nil {
				return errors.Wrap(err, "unable to bootstrap extension")
			}
		}
		if err = bootstrapPaths(); err != nil {
			return errors.Wrap(err, "unable to bootstrap paths")
		}
		if err = bootstrapLabel(); err != nil {
			return errors.Wrap(err, "unable to bootstrap label (%w)")
		}
		if m, err = bootstrapModel(); err != nil {
			return errors.Wrap(err, "unable to bootstrap binding (%w)")
		}
		for _, ext := range m.BootstrapExtensions {
			if err := ext.Bootstrap(m); err != nil {
				return errors.Wrap(err, "unable to bootstrap model-specific extension")
			}
		}

	} else {
		logrus.Warnf("no active instance")
	}
	return nil
}

func BootstrapBindings() error {
	var err error
	if err = loadBindings(); err != nil {
		return errors.Wrap(err, "unable to bootstrap config")
	}
	return nil
}

type BootstrapExtension interface {
	Bootstrap(m *Model) error
}

func bootstrapModel() (*Model, error) {
	l := GetLabel()
	if l != nil {
		m, found := GetModel(l.Model)
		if !found {
			return nil, errors.Errorf("no such model [%s]", l.Model)
		}

		if m.Parent != nil {
			if err := m.Merge(m.Parent); err != nil {
				return nil, errors.Wrap(err, "error merging parent")
			}
		}

		for _, factory := range m.ModelFactories {
			if err := factory.Build(m); err != nil {
				return nil, errors.Wrapf(err, "error executing factory [%s]", reflect.TypeOf(factory))
			}
		}

		m.BindLabel(l)
		if err := m.BindBindings(bindings); err != nil {
			return nil, errors.Wrap(err, "error bootstrapping model")
		}

		for _, factory := range m.Factories {
			if err := factory.Build(m); err != nil {
				return nil, errors.Wrapf(err, "error executing factory [%s]", reflect.TypeOf(factory))
			}
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
