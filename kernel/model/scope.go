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

type Scope struct {
	Variables Variables
	Data      Data
	Tags      Tags
	bound     bool
}

type Data map[string]interface{}
type Tags []string

type Variable struct {
	Description    string
	Default        interface{}
	Required       bool
	Scoped         bool
	GlobalFallback bool
	Binder         func(v *Variable, i interface{}, path ...string)
	Value          interface{}
	bound          bool
}

type Variables map[interface{}]interface{}

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
