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
	"strings"
)

const (
	InheritTagPrefix = "^"
)

type Scope struct {
	parent    *Scope
	Variables Variables
	Data      Data
	Tags      Tags
	bound     bool
}

func (scope *Scope) CloneScope() *Scope {
	result := &Scope{
		parent: scope.parent,
		bound:  scope.bound,
		Data:   Data{},
	}

	result.Variables = scope.Variables.Clone()

	for k, v := range scope.Data {
		result.Data[k] = v
	}

	for _, tag := range scope.Tags {
		result.Tags = append(result.Tags, tag)
	}

	return result
}

func (scope *Scope) Templatize(templater *Templater) {
	scope.Variables.ForEach(func(v *Variable) {
		v.Templatize(templater)
	})

	var newTags Tags
	for _, tag := range scope.Tags {
		newTags = append(newTags, templater.Templatize(tag))
	}
	scope.Tags = newTags
}

func (scope *Scope) setParent(parent *Scope) {
	scope.parent = parent

	tags := map[string]struct{}{}
	for _, tag := range scope.Tags {
		tags[tag] = struct{}{}
	}

	for _, tag := range parent.Tags {
		if strings.HasPrefix(tag, InheritTagPrefix) {
			tags[tag] = struct{}{}
		}
	}

	scope.Tags = nil
	for tag := range tags {
		scope.Tags = append(scope.Tags, tag)
	}
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

func (v *Variable) Templatize(templater *Templater) {
	if str, ok := v.Default.(string); ok {
		v.Default = templater.Templatize(str)
	}

	if str, ok := v.Value.(string); ok {
		v.Value = templater.Templatize(str)
	}
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

func (v Variables) NewVariable(name ...string) *Variable {
	inputMap := v
	for i := 0; i < (len(name) - 1); i++ {
		key := name[i]
		var nextMap Variables
		if value, found := inputMap[key]; found {
			var ok bool
			nextMap, ok = value.(Variables)
			if !ok {
				nextMap = Variables{}
				inputMap[key] = nextMap
			}
		} else {
			nextMap = Variables{}
			inputMap[key] = nextMap
		}
		inputMap = nextMap
	}

	result := &Variable{}
	inputMap[name[len(name)-1]] = result
	return result
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

func (v Variables) Clone() Variables {
	result := Variables{}
	for key, val := range v {
		switch tv := val.(type) {
		case Variables:
			result[key] = tv.Clone()
		case *Variable:
			varCopy := *tv
			result[key] = &varCopy
		default:
			result[key] = val
		}
	}
	return result
}

func (v Variables) ForEach(f func(v *Variable)) {
	for _, val := range v {
		switch tv := val.(type) {
		case Variables:
			tv.ForEach(f)
		case *Variable:
			f(tv)
		default:
		}
	}
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
