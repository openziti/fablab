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
	"strings"
)

func (m *Model) IsBound() bool {
	return m.bound
}

func (m *Model) Variable(name ...string) interface{} {
	value, found := m.GetVariable(name...)
	if !found {
		logrus.Fatalf("missing model variable [%v]", name)
	}
	return value
}

func (m *Model) GetVariable(name ...string) (interface{}, bool) {
	if len(name) < 1 {
		return nil, false
	}

	inputMap := m.Variables
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

func (m *Model) MustVariable(name ...string) interface{} {
	value, found := m.GetVariable(name...)
	if !found {
		logrus.Fatalf("missing data [%s]", name)
	}
	return value
}

func (m *Model) GetAction(name string) (Action, bool) {
	action, found := m.actions[name]
	return action, found
}

func (m *Model) GetAllHosts() []*Host {
	var hosts []*Host
	for _, r := range m.Regions {
		for _, h := range r.Hosts {
			hosts = append(hosts, h)
		}
	}
	return hosts
}

func (m *Model) GetHosts(regionSpec, hostSpec string) []*Host {
	var regions []*Region
	if strings.HasPrefix(regionSpec, "@") {
		regions = m.GetRegionsByTag(strings.TrimPrefix(regionSpec, "@"))
	} else {
		region := m.GetRegion(regionSpec)
		if region != nil {
			regions = append(regions, region)
		}
	}

	var hosts []*Host
	for _, region := range regions {
		if strings.HasPrefix(hostSpec, "@") {
			hosts = append(hosts, region.GetHostsByTag(strings.TrimPrefix(hostSpec, "@"))...)
		} else {
			host := region.GetHost(hostSpec)
			if host != nil {
				hosts = append(hosts, host)
			}
		}
	}

	return hosts
}

func (m *Model) GetHostByTags(regionTag, hostTag string) *Host {
	for regionId, region := range m.Regions {
		for _, tag := range region.Tags {
			if tag == regionTag {
				for hostId, host := range region.Hosts {
					for _, tag := range host.Tags {
						if tag == hostTag {
							logrus.Debugf("using [%s/%s] for tags [%s/%s]", regionId, hostId, regionTag, hostTag)
							return host
						}
					}
				}
			}
		}
	}
	logrus.Warnf("no resolution for tags [%s/%s]", regionTag, hostTag)
	return nil
}

func (m *Model) GetHostById(selectId string) (*Host, error) {
	hosts := make([]*Host, 0)
	for _, region := range m.Regions {
		for hostId, host := range region.Hosts {
			if hostId == selectId {
				hosts = append(hosts, host)
			}
		}
	}
	if len(hosts) != 1 {
		return nil, fmt.Errorf("found [%d] hosts with id [%s]", len(hosts), selectId)
	}
	return hosts[0], nil
}

func (m *Model) GetComponentsByTag(componentTag string) []*Component {
	var components []*Component
	for _, region := range m.Regions {
		for _, host := range region.Hosts {
			for _, component := range host.Components {
				for _, tag := range component.Tags {
					if tag == componentTag {
						components = append(components, component)
					}
				}
			}
		}
	}
	return components
}

func (m *Model) GetRegion(regionId string) *Region {
	region, found := m.Regions[regionId]
	if found {
		return region
	}
	return nil
}

func (m *Model) GetRegionByTag(regionTag string) *Region {
	for _, region := range m.Regions {
		for _, tag := range region.Tags {
			if tag == regionTag {
				return region
			}
		}
	}
	return nil
}

func (m *Model) GetRegionsByTag(regionTag string) []*Region {
	var regions []*Region
	for _, region := range m.Regions {
		for _, tag := range region.Tags {
			if tag == regionTag {
				regions = append(regions, region)
			}
		}
	}
	return regions
}

func (r *Region) GetHost(hostId string) *Host {
	host, found := r.Hosts[hostId]
	if found {
		return host
	}
	return nil
}

func (r *Region) GetHostsByTag(hostTag string) []*Host {
	var hosts []*Host
	for _, host := range r.Hosts {
		for _, tag := range host.Tags {
			if tag == hostTag {
				hosts = append(hosts, host)
			}
		}
	}
	return hosts
}

func (h *Host) GetComponents(componentSpec string) []*Component {
	var components []*Component
	if strings.HasPrefix(componentSpec, "@") {
		components = h.GetComponentsByTag(strings.TrimPrefix(componentSpec, "@"))
	} else {
		component := h.GetComponent(componentSpec)
		if component != nil {
			components = append(components, component)
		}
	}
	return components
}

func (h *Host) GetComponent(componentId string) *Component {
	component, found := h.Components[componentId]
	if found {
		return component
	}
	return nil
}

func (h *Host) GetComponentsByTag(componentTag string) []*Component {
	var components []*Component
	for _, component := range h.Components {
		for _, tag := range component.Tags {
			if tag == componentTag {
				components = append(components, component)
			}
		}
	}
	return components
}
