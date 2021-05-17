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

package operation

import (
	"fmt"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func TcpdumpCloser(host string) model.OperatingStage {
	return &tcpdumpCloser{
		host: host,
	}
}

func (t *tcpdumpCloser) Operate(run model.Run) error {
	m := run.GetModel()
	host, err := m.SelectHost(t.host)
	if err != nil {
		return err
	}

	ssh := lib.NewSshConfigFactoryImpl(m, host.PublicIp)

	if err := lib.RemoteKillFilter(ssh, "tcpdump", "sudo"); err != nil {
		return fmt.Errorf("error closing tcpdump (%w)", err)
	}
	logrus.Infof("tcpdump closed")
	return nil
}

type tcpdumpCloser struct {
	host string
}
