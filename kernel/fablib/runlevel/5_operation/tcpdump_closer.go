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
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func TcpdumpCloser(region, host string) model.OperatingStage {
	return &tcpdumpCloser{
		region: region,
		host:   host,
	}
}

func (t *tcpdumpCloser) Operate(m *model.Model, _ string) error {
	hosts := m.GetHosts(t.region, t.host)
	var ssh fablib.SshConfigFactory
	if len(hosts) == 1 {
		ssh = fablib.NewSshConfigFactoryImpl(m, hosts[0].PublicIp)
	} else {
		return fmt.Errorf("found [%d] hosts", len(hosts))
	}

	if err := fablib.RemoteKillFilter(ssh, "tcpdump", "sudo"); err != nil {
		return fmt.Errorf("error closing tcpdump (%w)", err)
	}
	logrus.Infof("tcpdump closed")
	return nil
}

type tcpdumpCloser struct {
	region string
	host   string
}
