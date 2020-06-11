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

package zitilib_characterization

import (
	"github.com/openziti/fablab/kernel/model"
)

func newHostsFactory() *hostsFactory {
	return &hostsFactory{}
}

func (f *hostsFactory) Build(m *model.Model) error {
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

	l := model.GetLabel()
	if l.Has("instance_type") {
		instanceType := l.Must("instance_type")
		for _, host := range m.GetAllHosts() {
			host.InstanceType = instanceType.(string)
		}
	}

	if l.Has("instance_resource_type") {
		instanceResourceType := l.Must("instance_resource_type")
		for _, host := range m.GetAllHosts() {
			host.InstanceResourceType = instanceResourceType.(string)
		}
	}

	if l.Has("spot_price") {
		spotPrice := l.Must("spot_price")
		for _, host := range m.GetAllHosts() {
			host.SpotPrice = spotPrice.(string)
		}
	}

	if l.Has("spot_type") {
		spotType := l.Must("spot_type")
		for _, host := range m.GetAllHosts() {
			host.SpotType = spotType.(string)
		}
	}

	return nil
}

type hostsFactory struct{}
