/*
	(c) Copyright NetFoundry, Inc.

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

package dilithium_actions

import (
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
)

func Clean() model.Action {
	return &clean{}
}

type clean struct{}

func (self *clean) Execute(m *model.Model) error {
	for _, host := range m.SelectHosts("*") {
		ssh := fablib.NewSshConfigFactoryImpl(m, host.PublicIp)
		if err := self.forHost(ssh); err != nil {
			return errors.Wrapf(err, "error cleaning host [%s/%s]", host.GetRegion().GetId(), host.GetId())
		}
	}
	return nil
}

func (self *clean) forHost(ssh fablib.SshConfigFactory) error {
	fis, err := fablib.RemoteFileList(ssh, ".")
	if err != nil {
		return errors.Wrap(err, "error retrieving files")
	}
	hasLogs := false
	for _, fi := range fis {
		if fi.Name() == "logs" && fi.IsDir() {
			hasLogs = true
			break
		}
	}
	if hasLogs {
		if _, err := fablib.RemoteExec(ssh, "rm -rf logs"); err != nil {
			return errors.Wrap(err, "error removing logs")
		}
		if _, err := fablib.RemoteExec(ssh, "mkdir logs"); err != nil {
			return errors.Wrap(err, "error re-creating logs")
		}
	}
	return nil
}
