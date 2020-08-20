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

package model

import (
	"fmt"
	"reflect"
)

func (m *Model) Dump() *Dump {
	return &Dump{
		Scope:   dumpScope(m.Scope),
		Regions: dumpRegions(m.Regions),
	}
}

func dumpScope(s Scope) *ScopeDump {
	dump := &ScopeDump{}
	empty := true
	if s.Variables != nil {
		variables := dumpVariables(s.Variables)
		dump.Variables = variables
		empty = false
	}
	if s.Data != nil {
		dump.Data = s.Data
		empty = false
	}
	if s.Tags != nil {
		dump.Tags = s.Tags
		empty = false
	}
	if !empty {
		return dump
	}
	return nil
}

func dumpVariables(vs Variables) map[string]interface{} {
	dump := make(map[string]interface{})
	for k, v := range vs {
		kk := fmt.Sprintf("%v", k)
		if vvs, ok := v.(Variables); ok {
			dump[kk] = dumpVariables(vvs)

		} else if vv, ok := v.(*Variable); ok {
			dump[kk] = dumpVariable(vv)

		} else {
			dump[kk] = reflect.TypeOf(v).String()
		}
	}
	return dump
}

func dumpVariable(v *Variable) *VariableDump {
	dump := &VariableDump{
		Description:    v.Description,
		Required:       v.Required,
		Scoped:         v.Scoped,
		GlobalFallback: v.GlobalFallback,
		Sensitive:      v.Sensitive,
		Bound:          v.bound,
	}
	if v.Default != nil {
		dump.Default = fmt.Sprintf("%v", v.Default)
	}
	if v.Binder != nil {
		dump.Binder = fmt.Sprintf("%p", v.Binder)
	}
	if v.Value != nil && !v.Sensitive {
		dump.Value = fmt.Sprintf("%v", v.Value)
	}
	return dump
}

func dumpRegions(rs map[string]*Region) map[string]*RegionDump {
	dumps := make(map[string]*RegionDump)
	for k, v := range rs {
		dumps[k] = dumpRegion(v)
	}
	return dumps
}

func dumpRegion(r *Region) *RegionDump {
	return &RegionDump{
		Scope: dumpScope(r.Scope),
		Id:    r.Id,
		Az:    r.Az,
		Hosts: dumpHosts(r.Hosts),
	}
}

func dumpHosts(hs map[string]*Host) map[string]*HostDump {
	dumps := make(map[string]*HostDump)
	for k, v := range hs {
		dumps[k] = DumpHost(v)
	}
	return dumps
}

func DumpHost(h *Host) *HostDump {
	return &HostDump{
		Scope:                dumpScope(h.Scope),
		PublicIp:             h.PublicIp,
		PrivateIp:            h.PrivateIp,
		InstanceType:         h.InstanceType,
		InstanceResourceType: h.InstanceResourceType,
		SpotPrice:            h.SpotPrice,
		SpotType:             h.SpotType,
	}
}

func dumpComponents(cs map[string]*Component) map[string]*ComponentDump {
	dumps := make(map[string]*ComponentDump)
	for k, v := range cs {
		dumps[k] = dumpComponent(v)
	}
	return dumps
}

func dumpComponent(c *Component) *ComponentDump {
	return &ComponentDump{
		Scope:           dumpScope(c.Scope),
		ScriptSrc:       c.ScriptSrc,
		ScriptName:      c.ScriptName,
		ConfigSrc:       c.ConfigSrc,
		ConfigName:      c.ConfigName,
		BinaryName:      c.BinaryName,
		PublicIdentity:  c.PublicIdentity,
		PrivateIdentity: c.PrivateIdentity,
	}
}

type Dump struct {
	Scope   *ScopeDump             `json:"scope,omitempty"`
	Regions map[string]*RegionDump `json:"regions"`
}

type ScopeDump struct {
	Variables map[string]interface{} `json:"variables,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
}

type VariableDump struct {
	Description    string `json:"description,omitempty"`
	Default        string `json:"default,omitempty"`
	Required       bool   `json:"required"`
	Scoped         bool   `json:"scoped"`
	GlobalFallback bool   `json:"global_fallback"`
	Sensitive      bool   `json:"sensitive"`
	Binder         string `json:"binder,omitempty"`
	Value          string `json:"value,omitempty"`
	Bound          bool   `json:"bound"`
}

type RegionDump struct {
	Scope *ScopeDump           `json:"scope,omitempty"`
	Id    string               `json:"id,omitempty"`
	Az    string               `json:"az,omitempty"`
	Hosts map[string]*HostDump `json:"hosts,omitempty"`
}

type HostDump struct {
	Scope                *ScopeDump                `json:"scope,omitempty"`
	PublicIp             string                    `json:"public_ip,omitempty"`
	PrivateIp            string                    `json:"private_ip,omitempty"`
	InstanceType         string                    `json:"instance_type,omitempty"`
	InstanceResourceType string                    `json:"instance_resource_type,omitempty"`
	SpotPrice            string                    `json:"spot_price,omitempty"`
	SpotType             string                    `json:"spot_type,omitempty"`
	Components           map[string]*ComponentDump `json:"components,omitempty"`
}

type ComponentDump struct {
	Scope           *ScopeDump `json:"scope,omitempty"`
	ScriptSrc       string     `json:"script_src,omitempty"`
	ScriptName      string     `json:"script_name,omitempty"`
	ConfigSrc       string     `json:"config_src,omitempty"`
	ConfigName      string     `json:"config_name,omitempty"`
	BinaryName      string     `json:"binary_name,omitempty"`
	PublicIdentity  string     `json:"public_identity,omitempty`
	PrivateIdentity string     `json:"private_identity,omitempty"`
}
