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

package component

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/model"
)

func Start(regionSpec, hostSpec, componentSpec string) model.Action {
	return &start{
		regionSpec:    regionSpec,
		hostSpec:      hostSpec,
		componentSpec: componentSpec,
	}
}

func (start *start) Execute(m *model.Model) error {
	hosts := m.GetHosts(start.regionSpec, start.hostSpec)
	for _, h := range hosts {
		components := h.GetComponents(start.componentSpec)
		for _, c := range components {
			sshConfigFactory := fablib.NewSshConfigFactoryImpl(m, h.PublicIp)

			if err := fablib.LaunchService(sshConfigFactory, c.BinaryName, c.ConfigName); err != nil {
				return fmt.Errorf("error starting component [%s] on [%s] (%s)", c.BinaryName, h.PublicIp, err)
			}
		}
	}
	return nil
}

type start struct {
	regionSpec    string
	hostSpec      string
	componentSpec string
}
