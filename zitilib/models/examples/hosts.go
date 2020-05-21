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

func newHostsFactory() *hostsFactory {
	return &hostsFactory{}
}

func (f *hostsFactory) Build(m *model.Model) error {
	for _, host := range m.GetAllHosts() {
		host.InstanceType = "t2.micro"
	}

	v, found := m.GetVariable("instance_type")
	if found {
		instanceType := v.(string)
		for _, host := range m.GetAllHosts() {
			host.InstanceType = instanceType
		}
	}

	return nil
}

type hostsFactory struct{}
