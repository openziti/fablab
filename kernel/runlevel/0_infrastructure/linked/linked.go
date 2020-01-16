/*
	Copyright 2020 NetFoundry, Inc.

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

package linked_0

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Linked() model.InfrastructureStage {
	return &linked{}
}

func (linked *linked) Express(m *model.Model, l *model.Label) error {
	var parent string
	if value, found := l.Bindings["parent"]; found {
		parent = value
	} else {
		return fmt.Errorf("missing 'parent' label binding")
	}

	parentL, err := model.LoadLabelForInstance(parent)
	if err != nil {
		return fmt.Errorf("error loading parent label [%s] (%w)", parent, err)
	}

	for k, v := range parentL.Bindings {
		if k != "parent" {
			l.Bindings[k] = v
			logrus.Infof("copied [%s]=[%s] from parent", k, v)
		}
	}

	if err := l.Save(); err != nil {
		return fmt.Errorf("error saving updated label (%w)", err)
	}

	m.BindLabel(l)
	return nil
}

type linked struct{}
