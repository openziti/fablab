/*
	(c) Copyright NetFoundry Inc. Inc.

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
	"github.com/openziti/fablab/kernel/libssh"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Exec(h *model.Host, cmds ...string) model.Action {
	return &exec{
		h:    h,
		cmds: cmds,
	}
}

func (exec *exec) Execute(model.Run) error {
	sshConfigFactory := exec.h.NewSshConfigFactory()

	if o, err := libssh.RemoteExecAll(sshConfigFactory, exec.cmds...); err != nil {
		logrus.Errorf("output [%s]", o)
		return fmt.Errorf("error executing process on [%s] (%s)", exec.h.PublicIp, err)
	}
	return nil
}

type exec struct {
	h    *model.Host
	cmds []string
}
