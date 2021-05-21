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
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
)

func GroupKill(hostSpec, match string) model.Action {
	return &groupKill{
		hostSpec: hostSpec,
		match:    match,
	}
}

func (groupKill *groupKill) Execute(m *model.Model) error {
	for _, h := range m.SelectHosts(groupKill.hostSpec) {

		sshConfigFactory := lib.NewSshConfigFactoryImpl(h)
		if err := lib.RemoteKill(sshConfigFactory, groupKill.match); err != nil {
			return fmt.Errorf("error killing [%s] on [%s] (%s)", groupKill.match, h.PublicIp, err)
		}
	}
	return nil
}

type groupKill struct {
	hostSpec string
	match    string
}
