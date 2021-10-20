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
	if model == nil {
		return errors.New("no model initialized, exiting")
	}
	if model.Id == "" {
		return errors.New("model id not set, exiting")
	}

	model.init()

	if err = BootstrapInstance(); err != nil {
		return errors.Wrap(err, "unable to bootstrap instance config")
	}

	if err = BootstrapBindings(); err != nil {
		return errors.Wrap(err, "unable to bootstrap config")
	}
	model.VarConfig.BindingResolver.UpdateVariables(bindings)

	for _, ext := range bootstrapExtensions {
		if err := ext.Bootstrap(model); err != nil {
			return errors.Wrap(err, "unable to bootstrap extension")
		}
	}
	if err = bootstrapLabel(); err != nil {
		return errors.Wrap(err, "unable to bootstrap label (%w)")
	}
	model.VarConfig.LabelResolver.UpdateVariables(label.Bindings)

	if err = bootstrapModel(); err != nil {
		return errors.Wrap(err, "unable to bootstrap binding (%w)")
	}
	for _, ext := range model.BootstrapExtensions {
		if err := ext.Bootstrap(model); err != nil {
			return errors.Wrap(err, "unable to bootstrap model-specific extension")
		}
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

func bootstrapModel() error {
	l := GetLabel()
	if l != nil {
		if l.Model != model.GetId() {
			return errors.Errorf("running model '%v' doesn't match project workspace model '%v'", model.GetId(), l.Model)
		}

		for _, factory := range model.StructureFactories {
			if err := factory.Build(model); err != nil {
				return errors.Wrapf(err, "error executing factory [%s]", reflect.TypeOf(factory))
			}
		}

		model.BindLabel(l)

		for _, factory := range model.Factories {
			if err := factory.Build(model); err != nil {
				return errors.Wrapf(err, "error executing factory [%s]", reflect.TypeOf(factory))
			}
		}

		model.actions = make(map[string]Action)
		for name, binder := range model.Actions {
			model.actions[name] = binder(model)
			logrus.Debugf("bound action [%s]", name)
		}

		return nil

	} else {
		logrus.Warn("no run label found")
	}
	return nil
}
