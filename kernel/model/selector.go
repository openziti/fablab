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
	"github.com/openziti/foundation/util/stringz"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

const (
	SelectorTagPrefix = "."
	SelectorIdPrefox  = "#"
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

func (m *Model) GetAction(name string) (Action, bool) {
	action, found := m.actions[name]
	return action, found
}

func (m *Model) SelectRegions(spec string) []*Region {
	matcher := compileEntityMatcher(spec, 1, "region")
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
	matcher := compileEntityMatcher(spec, 2, "host")
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
	matcher := compileEntityMatcher(spec, 3, "component")
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

func compileEntityMatcher(in string, maxDepth uint8, entityType string) EntityMatcher {
	parts := strings.Split(in, ">")
	if len(parts) > int(maxDepth) {
		panic(errors.Errorf("invalid %v spec '%v', only %v level(s) may be specified", entityType, in, maxDepth))
	}
	matchers := specsToMatchers(parts)

	if len(matchers) == 1 {
		return func(entity Entity) bool {
			return matchers[0](entity)
		}
	}

	if len(matchers) == 2 {
		return func(entity Entity) bool {
			return matchers[0](entity.GetParentEntity()) && matchers[1](entity)
		}
	}

	return func(entity Entity) bool {
		return matchers[0](entity.GetParentEntity().GetParentEntity()) && matchers[1](entity.GetParentEntity()) && matchers[2](entity)
	}
}

func compileSpec(in string) EntityMatcher {
	specs := strings.Split(in, ",")
	result := specToMatcher(specs[0])
	for _, spec := range specs[1:] {
		result = result.Or(specToMatcher(spec))
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

	if strings.HasPrefix(spec, SelectorTagPrefix) {
		tags := strings.Split(spec, SelectorTagPrefix)
		tags = stringz.Remove(tags, "")
		result := newTagSelector(tags[0])
		for _, tag := range tags[1:] {
			result = result.And(newTagSelector(tag))
		}
		return result
	}

	if strings.HasPrefix(spec, SelectorIdPrefox) {
		id := strings.TrimPrefix(spec, SelectorIdPrefox)
		return func(e Entity) bool {
			return id == e.GetId()
		}
	}

	panic(errors.Errorf("invalid selector '%v', not a .tag #id or ", spec))
}

func newTagSelector(tag string) EntityMatcher {
	return func(e Entity) bool {
		return stringz.Contains(e.GetScope().Tags, tag)
	}
}

func specsToMatchers(parts []string) []EntityMatcher {
	var result []EntityMatcher
	for _, part := range parts {
		result = append(result, compileSpec(part))
	}
	return result
}

func Selector(levels ...string) string {
	return strings.Join(levels, " > ")
}
