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

package __operation

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func LoopDialer(host *model.Host, id, scenario, endpoint string, joiner chan struct{}) model.OperatingStage {
	return &loopDialer{
		host:     host,
		id:       id,
		scenario: scenario,
		endpoint: endpoint,
		joiner:   joiner,
	}
}

func (self *loopDialer) Operate(m *model.Model, run string) error {
	ssh := fablib.NewSshConfigFactoryImpl(m, self.host.PublicIp)
	if err := fablib.RemoteKill(ssh, "ziti-fabric-test loop2 dialer"); err != nil {
		return fmt.Errorf("error killing loop2 listeners (%w)", err)
	}

	go self.run(m, run)
	return nil
}

func (self *loopDialer) run(m *model.Model, run string) {
	defer func() {
		if self.joiner != nil {
			close(self.joiner)
			logrus.Debugf("closed joiner")
		}
	}()

	ssh := fablib.NewSshConfigFactoryImpl(m, self.host.PublicIp)
	logFile := fmt.Sprintf("/home/%s/logs/loop2-dialer-%s.log", ssh.User(), run)
	dialerCmd := fmt.Sprintf("/home/%s/fablab/bin/ziti-fabric-test loop2 dialer /home/%s/fablab/cfg/%s -e %s -s %s >> %s 2>&1", ssh.User(), ssh.User(), self.scenario, self.endpoint, self.id, logFile)
	if output, err := fablib.RemoteExec(ssh, dialerCmd); err != nil {
		logrus.Errorf("error starting loop dialer [%s] (%w)", output, err)
	}
}

type loopDialer struct {
	host     *model.Host
	id       string
	endpoint string
	scenario string
	joiner   chan struct{}
}
