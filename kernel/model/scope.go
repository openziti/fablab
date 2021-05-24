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
	"sort"
)

type Scope struct {
	parent    *Scope
	Variables Variables
	Data      Data
	Tags      Tags
	bound     bool
}

func (scope *Scope) setParent(parent *Scope) {
	scope.parent = parent
	sort.Strings(scope.Tags)
}

func (scope *Scope) HasTag(tag string) bool {
	for _, hostTag := range scope.Tags {
		if hostTag == tag {
			return true
		}
	}
	return false
}

func (scope *Scope) WithTags(tags ...string) *Scope {
	scope.Tags = tags
	return scope
}

type Data map[string]interface{}
type Tags []string

type Variable struct {
	Description    string
	Default        interface{}
	Required       bool
	Scoped         bool
	GlobalFallback bool
	Sensitive      bool
	Binder         func(v *Variable, i interface{}, path ...string)
	Value          interface{}
	bound          bool
}

type Variables map[interface{}]interface{}

func (v Variables) Put(newValue interface{}, name ...string) error {
	if len(name) < 1 {
		return errors.New("empty name")
	}

	inputMap := v
	for i := 0; i < (len(name) - 1); i++ {
		key := name[i]
		if value, found := inputMap[key]; found {
			lowerMap, ok := value.(Variables)
			if !ok {
				return errors.Errorf("invalid path type [%s]", key)
			}
			inputMap = lowerMap
		}
	}

	value, found := inputMap[name[len(name)-1]]
	if found {
		variable, ok := value.(*Variable)
		if !ok {
			return errors.Errorf("path not variable leaf")
		}
		variable.Value = newValue
		variable.bound = true
	}

	return nil
}

func (v Variables) Get(name ...string) (interface{}, bool) {
	if len(name) < 1 {
		return nil, false
	}

	inputMap := v
	for i := 0; i < (len(name) - 1); i++ {
		key := name[i]
		if value, found := inputMap[key]; found {
			lowerMap, ok := value.(Variables)
			if !ok {
				return nil, false
			}
			inputMap = lowerMap
		}
	}

	value, found := inputMap[name[len(name)-1]]
	if found {
		variable, ok := value.(*Variable)
		if !ok {
			return nil, false
		}
		if variable.Required {
			if !variable.bound {
				logrus.Fatalf("required variable %v missing", name)
			}
			return variable.Value, true
		} else {
			if variable.bound {
				return variable.Value, true
			} else {
				return variable.Default, true
			}
		}
	}
	return nil, false
}

func (v Variables) Must(name ...string) interface{} {
	value, found := v.Get(name...)
	if !found {
		logrus.Fatalf("missing variable [%s]", name)
	}
	return value
}

func (m *Model) IterateScopes(f func(i interface{}, path ...string)) {
	f(m, []string{}...)
	for regionId, r := range m.Regions {
		f(r, []string{regionId}...)
		for hostId, h := range r.Hosts {
			f(h, []string{regionId, hostId}...)
			for componentId, c := range h.Components {
				f(c, []string{regionId, hostId, componentId}...)
			}
		}
	}
}
