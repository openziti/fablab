/*
	Copyright 2020 Netfoundry, Inc.

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
	"github.com/jinzhu/copier"
)

func (m *Model) Merge(parent *Model) error {
	var err error
	if m.Scope, err = m.Scope.Merge(parent.Scope); err != nil {
		return fmt.Errorf("error merging model scope (%w)", err)
	}
	if m.Regions, err = m.Regions.Merge(parent.Regions); err != nil {
		return fmt.Errorf("error merging model regions (%w)", err)
	}
	return nil
}

func (s Scope) Merge(parent Scope) (Scope, error) {
	merged := Scope{}
	if err := copier.Copy(&merged, parent); err != nil {
		return Scope{}, fmt.Errorf("error copying parent (%w)", err)
	}

	for k, v := range s.Variables {
		merged.Variables[k] = v
	}

	for k, v := range s.Data {
		merged.Data[k] = v
	}

	for k, v := range s.Tags {
		merged.Tags[k] = v
	}

	return merged, nil
}

func (r Regions) Merge(parent Regions) (Regions, error) {
	merged := make(Regions)
	if err := copier.Copy(&merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	for k, v := range r {
		if pv, found := merged[k]; found {
			if mergedRegion, err := v.Merge(pv); err == nil {
				merged[k] = mergedRegion
			} else {
				return nil, fmt.Errorf("error merging region [%s] (%w)", k, err)
			}
		} else {
			merged[k] = v
		}
	}

	return merged, nil
}

func (r *Region) Merge(parent *Region) (*Region, error) {
	merged := &Region{}
	if err := copier.Copy(merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	var err error
	if merged.Scope, err = r.Scope.Merge(parent.Scope); err != nil {
		return nil, fmt.Errorf("error merging region scope (%w)", err)
	}

	if r.Id != "" {
		merged.Id = r.Id
	}
	if r.Az != "" {
		merged.Az = r.Az
	}


	if merged.Hosts, err = r.Hosts.Merge(parent.Hosts); err != nil {
		return nil, fmt.Errorf("error merging hosts (%w)", err)
	}

	return merged, nil
}

func (h Hosts) Merge(parent Hosts) (Hosts, error) {
	merged := make(Hosts)
	if err := copier.Copy(&merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	for k, v := range h {
		if pv, found := merged[k]; found {
			if mergedHost, err := v.Merge(pv); err == nil {
				merged[k] = mergedHost
			} else {
				return nil, fmt.Errorf("error merging host (%w)", err)
			}
		} else {
			merged[k] = v
		}
	}

	return nil, nil
}

func (h *Host) Merge(parent *Host) (*Host, error) {
	merged := &Host{}
	if err := copier.Copy(merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	var err error
	if merged.Scope, err = h.Scope.Merge(parent.Scope); err != nil {
		return nil, fmt.Errorf("error merging host scope (%w)", err)
	}

	if h.PublicIp != "" {
		merged.PublicIp = h.PublicIp
	}
	if h.PrivateIp != "" {
		merged.PrivateIp = h.PrivateIp
	}
	if h.InstanceType != "" {
		merged.InstanceType = h.InstanceType
	}

	if merged.Components, err = h.Components.Merge(parent.Components); err != nil {
		return nil, fmt.Errorf("error merging components (%w)", err)
	}

	return merged, nil
}

func (c Components) Merge(parent Components) (Components, error) {
	merged := make(Components)
	if err := copier.Copy(&merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	for k, v := range c {
		if pv, found := merged[k]; found {
			if mergedComponent, err := v.Merge(pv); err == nil {
				merged[k] = mergedComponent
			} else {
				return nil, fmt.Errorf("error merging component (%w)", err)
			}
		} else {
			merged[k] = v
		}
	}

	return merged, nil
}

func (c *Component) Merge(parent *Component) (*Component, error) {
	merged := &Component{}
	if err := copier.Copy(merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	var err error
	if merged.Scope, err = c.Scope.Merge(parent.Scope); err != nil {
		return nil, fmt.Errorf("error merging component scope (%w)", err)
	}

	if c.ScriptSrc != "" {
		merged.ScriptSrc = c.ScriptSrc
	}
	if c.ScriptName != "" {
		merged.ScriptName = c.ScriptName
	}
	if c.ConfigSrc != "" {
		merged.ConfigSrc = c.ConfigSrc
	}
	if c.ConfigName != "" {
		merged.ConfigName = c.ConfigName
	}
	if c.BinaryName != "" {
		merged.BinaryName = c.BinaryName
	}
	if c.PublicIdentity != "" {
		merged.PublicIdentity = c.PublicIdentity
	}
	if c.PrivateIdentity != "" {
		merged.PrivateIdentity = c.PrivateIdentity
	}

	return merged, nil
}