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

package host

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/internal"
	"github.com/netfoundry/fablab/kernel/model"
)

func GroupKill(regionSpec, hostSpec, match string) model.Action {
	return &groupKill{
		regionSpec: regionSpec,
		hostSpec:   hostSpec,
		match:      match,
	}
}

func (groupKill *groupKill) Execute(m *model.Model) error {
	hosts := m.GetHosts(groupKill.regionSpec, groupKill.hostSpec)
	for _, h := range hosts {
		sshUsername := m.MustVariable("credentials", "ssh", "username").(string)
		if err := internal.RemoteKill(sshUsername, h.PublicIp, groupKill.match); err != nil {
			return fmt.Errorf("error killing [%s] on [%s] (%s)", groupKill.match, h.PublicIp, err)
		}
	}
	return nil
}

type groupKill struct {
	regionSpec string
	hostSpec   string
	match      string
}
