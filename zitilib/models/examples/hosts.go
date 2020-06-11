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

package zitilib_examples

import "github.com/openziti/fablab/kernel/model"

func newHostsFactory() model.Factory {
	return &hostsFactory{}
}

func (_ *hostsFactory) Build(m *model.Model) error {
	for _, host := range m.GetAllHosts() {
		if host.InstanceType == "" {
			host.InstanceType = "t2.micro"
		}
		if host.InstanceResourceType == "" {
			host.InstanceResourceType = "ondemand"
		}
		if host.InstanceResourceType == "spot" {
			if host.SpotPrice == "" {
				host.SpotPrice = "0.02"
			}
			if host.SpotType == "" {
				host.SpotType = "one-time"
			}
		}
	}

	v, found := m.GetVariable("instance_type")
	if found {
		instanceType := v.(string)
		for _, host := range m.GetAllHosts() {
			host.InstanceType = instanceType
		}
	}

	v, found = m.GetVariable("instance_resource_type")
	if found {
		instanceResourceType := v.(string)
		for _, host := range m.GetAllHosts() {
			host.InstanceResourceType = instanceResourceType
		}
	}

	v, found = m.GetVariable("spot_price")
	if found {
		spotPrice := v.(string)
		for _, host := range m.GetAllHosts() {
			host.SpotPrice = spotPrice
		}
	}

	v, found = m.GetVariable("spot_type")
	if found {
		spotType := v.(string)
		for _, host := range m.GetAllHosts() {
			host.SpotType = spotType
		}
	}

	return nil
}

type hostsFactory struct{}
