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
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
)

func Stop(regionSpec, hostSpec, componentSpec string) model.Action {
	return &stop{
		regionSpec:    regionSpec,
		hostSpec:      hostSpec,
		componentSpec: componentSpec,
	}
}

func (stop *stop) Execute(m *model.Model) error {
	hosts := m.GetHosts(stop.regionSpec, stop.hostSpec)
	for _, h := range hosts {
		components := h.GetComponents(stop.componentSpec)
		for _, c := range components {
			sshUsername := m.MustVariable("credentials", "ssh", "username").(string)
			if err := fablib.KillService(sshUsername, h.PublicIp, c.BinaryName); err != nil {
				return fmt.Errorf("error stopping component [%s] on [%s] (%s)", c.BinaryName, h.PublicIp, err)
			}
		}
	}

	return nil
}

type stop struct {
	regionSpec    string
	hostSpec      string
	componentSpec string
}
