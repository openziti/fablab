/*
	Copyright 2019 Netfoundry, Inc.

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
	"github.com/netfoundry/fablab/kernel"
	"github.com/netfoundry/fablab/kernel/lib"
	"github.com/sirupsen/logrus"
)

func Exec(h *kernel.Host, cmd string) kernel.Action {
	return &exec{
		h:   h,
		cmd: cmd,
	}
}

func (exec *exec) Execute(m *kernel.Model) error {
	sshUsername := m.MustVariable("credentials", "ssh", "username").(string)
	if o, err := lib.RemoteExec(sshUsername, exec.h.PublicIp, exec.cmd); err != nil {
		logrus.Errorf("output [%s]", o)
		return fmt.Errorf("error executing process [%s] on [%s] (%s)", exec.cmd, exec.h.PublicIp, err)
	}
	return nil
}

type exec struct {
	h   *kernel.Host
	cmd string
}
