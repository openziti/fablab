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
	"github.com/openziti/fablab/kernel/fablib/parallel"
	"github.com/openziti/foundation/util/errorz"
	"github.com/openziti/foundation/util/stringz"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

const (
	SelectorTagPrefix = "."
	SelectorIdPrefix  = "#"
)

func (m *Model) IsBound() bool {
	return m.bound
}

func (m *Model) GetVariable(name ...string) (interface{}, bool) {
	return m.Variables.Get(name...)
}

func (m *Model) MustVariable(name ...string) interface{} {
	return m.Variables.Must(name...)
}

func (m *Model) MustStringVariable(name ...string) string {
	value, found := m.GetVariable(name...)
	if !found {
		logrus.Fatalf("missing variable [%s]", name)
	}
	result, ok := value.(string)
	if !ok {
		logrus.Fatalf("variable [%v] expected to have type string, but was %v", name, reflect.TypeOf(value))
	}
	return result
}

func (m *Model) GetRequiredStringVariable(holder errorz.ErrorHolder, name ...string) string {
	value, found := m.GetVariable(name...)
	if !found {
		holder.SetError(errors.Errorf("missing variable [%s]", name))
		return ""
	}
	result, ok := value.(string)
	if !ok {
		holder.SetError(errors.Errorf("variable [%v] expected to have type string, but was %v", name, reflect.TypeOf(value)))
	}
	return result
}

func (m *Model) GetAction(name string) (Action, bool) {
	action, found := m.actions[name]
	return action, found
}

func (m *Model) SelectRegions(spec string) []*Region {
	matcher := compileSelector(spec)
	var regions []*Region
	for _, region := range m.Regions {
		if matcher(region) {
			regions = append(regions, region)
		}
	}
	return regions
}

func (m *Model) SelectRegion(spec string) (*Region, error) {
	regions := m.SelectRegions(spec)
	if len(regions) == 1 {
		return regions[0], nil
	} else {
		return nil, errors.Errorf("[%s] matched [%d] regions, expected 1", spec, len(regions))
	}
}

func (m *Model) MustSelectRegion(spec string) *Region {
	region, err := m.SelectRegion(spec)
	if err != nil {
		panic(err)
	}
	return region
}

func (m *Model) SelectHosts(spec string) []*Host {
	matcher := compileSelector(spec)
	var hosts []*Host
	for _, region := range m.Regions {
		for _, host := range region.Hosts {
			if matcher(host) {
				hosts = append(hosts, host)
			}
		}
	}
	return hosts
}

func (m *Model) MustSelectHosts(spec string, minCount int) ([]*Host, error) {
	hosts := m.SelectHosts(spec)
	if len(hosts) < minCount {
		return nil, errors.Errorf("[%s] matched [%d] hosts, expected at least %v", spec, len(hosts), minCount)
	}
	return hosts, nil
}

func (m *Model) SelectHost(spec string) (*Host, error) {
	hosts := m.SelectHosts(spec)
	if len(hosts) == 1 {
		return hosts[0], nil
	} else {
		return nil, errors.Errorf("[%s] matched [%d] hosts, expected 1", spec, len(hosts))
	}
}

func (m *Model) MustSelectHost(spec string) *Host {
	host, err := m.SelectHost(spec)
	if err != nil {
		panic(err)
	}
	return host
}

func (m *Model) SelectComponents(spec string) []*Component {
	matcher := compileSelector(spec)
	var components []*Component
	for _, region := range m.Regions {
		for _, host := range region.Hosts {
			for _, component := range host.Components {
				if matcher(component) {
					components = append(components, component)
				}
			}
		}
	}
	return components
}

func (m *Model) SelectComponent(spec string) (*Component, error) {
	components := m.SelectComponents(spec)
	if len(components) == 1 {
		return components[0], nil
	} else {
		return nil, errors.Errorf("[%s] matched [%d] components, expected 1", spec, len(components))
	}
}

func (m *Model) ForEachHost(spec string, parallel bool, f func(host *Host) error) error {
	if parallel {
		return m.ForEachHostParallel(spec, f)
	}
	return m.ForEachHostSequential(spec, f)
}

func (m *Model) ForEachHostSequential(spec string, f func(host *Host) error) error {
	hosts := m.SelectHosts(spec)
	for _, host := range hosts {
		if err := f(host); err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) ForEachHostParallel(spec string, f func(host *Host) error) error {
	hosts := m.SelectHosts(spec)
	var tasks []parallel.Task
	for _, host := range hosts {
		boundHost := host
		tasks = append(tasks, func() error {
			return f(boundHost)
		})
	}
	return parallel.Execute(tasks)
}

func (m *Model) ForEachComponent(spec string, parallel bool, f func(c *Component) error) error {
	if parallel {
		return m.ForEachComponentParallel(spec, f)
	}
	return m.ForEachComponentSequential(spec, f)
}

func (m *Model) ForEachComponentSequential(spec string, f func(c *Component) error) error {
	components := m.SelectComponents(spec)
	for _, c := range components {
		if err := f(c); err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) ForEachComponentParallel(spec string, f func(c *Component) error) error {
	components := m.SelectComponents(spec)
	var tasks []parallel.Task
	for _, component := range components {
		boundComponent := component
		tasks = append(tasks, func() error {
			return f(boundComponent)
		})
	}
	return parallel.Execute(tasks)
}

type EntityMatcher func(Entity) bool

func (m EntityMatcher) Or(m2 EntityMatcher) EntityMatcher {
	return func(e Entity) bool {
		return m(e) || m2(e)
	}
}

func (m EntityMatcher) And(m2 EntityMatcher) EntityMatcher {
	return func(e Entity) bool {
		return m(e) && m2(e)
	}
}

func compileSelector(in string) EntityMatcher {
	parts := strings.Split(in, ">")

	// stack them in reverse order so we can evaluate target entity up the parents
	matchers := make([]EntityMatcher, len(parts))
	for idx, part := range parts {
		matcher := compileLevelSelector(part)
		matchers[len(parts)-(idx+1)] = matcher
	}

	return func(entity Entity) bool {
		current := entity
		for _, matcher := range matchers {
			if current == nil || !matcher(current) {
				return false
			}
			current = current.GetParentEntity()
		}
		return true
	}
}

func compileLevelSelector(in string) EntityMatcher {
	specs := strings.Split(in, ",")
	result := compileSelectorGroup(specs[0])
	for _, spec := range specs[1:] {
		result = result.Or(compileSelectorGroup(spec))
	}
	return result
}

func compileSelectorGroup(in string) EntityMatcher {
	parts := strings.Split(in, " ")
	parts = stringz.Remove(parts, "")
	result := specToMatcher(parts[0])
	for _, part := range parts[1:] {
		result = result.And(specToMatcher(part))
	}
	return result
}

func specToMatcher(spec string) EntityMatcher {
	spec = strings.TrimSpace(spec)
	if spec == "*" {
		return func(Entity) bool {
			return true
		}
	}

	var entityType string
	var entityId string
	var entityTags []string

	if !strings.HasPrefix(spec, SelectorTagPrefix) && !strings.HasPrefix(spec, SelectorIdPrefix) {
		if idx := strings.Index(spec, SelectorIdPrefix); idx > 0 {
			entityType = spec[0:idx]
			spec = spec[idx:]
		} else if idx := strings.Index(spec, SelectorTagPrefix); idx > 0 {
			entityType = spec[0:idx]
			spec = spec[idx:]
		}
	}

	if strings.HasPrefix(spec, SelectorIdPrefix) {
		if idx := strings.Index(spec, SelectorTagPrefix); idx > 0 {
			entityId = spec[1:idx]
			spec = spec[idx:]
		} else {
			entityId = spec[1:]
			spec = ""
		}
	}

	if strings.HasPrefix(spec, SelectorTagPrefix) {
		entityTags = strings.Split(spec, SelectorTagPrefix)
		entityTags = stringz.Remove(entityTags, "")
	}

	var matcher EntityMatcher
	if entityId != "" {
		matcher = func(e Entity) bool {
			return e.GetId() == entityId
		}
	}

	for _, tag := range entityTags {
		tagMatcher := newTagSelector(tag)
		if matcher == nil {
			matcher = tagMatcher
		} else {
			matcher = matcher.And(tagMatcher)
		}
	}

	if matcher == nil {
		return func(e Entity) bool {
			return true
		}
	}

	if entityType == "" {
		return matcher
	}

	return func(e Entity) bool {
		return e.Matches(entityType, matcher)
	}
}

func newTagSelector(tag string) EntityMatcher {
	return func(e Entity) bool {
		return stringz.Contains(e.GetScope().Tags, tag)
	}
}

func Selector(levels ...string) string {
	return strings.Join(levels, " > ")
}
