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

package dilithium

import (
	"github.com/openziti/fablab/kernel/model"
)

func newHostsFactory() model.Factory {
	return &hostsFactory{}
}

func (_ *hostsFactory) Build(m *model.Model) error {
	l := model.GetLabel()

	if l.Has("instance_type") {
		instanceType := l.Must("instance_type")
		for _, host := range m.SelectHosts("*", "*") {
			host.InstanceType = instanceType.(string)
		}
	}

	if l.Has("remote_region_id") {
		regionId := l.Must("remote_region_id")
		m.MustSelectRegion("remote").Id = regionId.(string)
	}
	if l.Has("remote_region_az") {
		regionAz := l.Must("remote_region_az")
		m.MustSelectRegion("remote").Az = regionAz.(string)
	}

	return nil
}

type hostsFactory struct{}
