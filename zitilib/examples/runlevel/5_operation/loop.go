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

package operation

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Loop() model.OperatingStage {
	return &loopOperation{}
}

func (self *loopOperation) Operate(m *model.Model, run string) error {
	listenerHosts := m.GetHosts("@terminator", "@loop-listener")
	if len(listenerHosts) < 1 {
		return fmt.Errorf("no loop listener hosts in model")
	}

	initiatorHost := m.GetHosts("@initiator", "@initiator")
	if len(initiatorHost) != 1 {
		return fmt.Errorf("expected 1 initiator host in model")
	}

	var dialerHosts []*model.Host
	var dialerIds []string
	for dialerId, dialerHost := range m.GetRegionByTag("initiator").Hosts {
		if dialerHost.HasTag("loop-dialer") {
			dialerHosts = append(dialerHosts, dialerHost)
			dialerIds = append(dialerIds, dialerId)
		}
	}
	if len(dialerHosts) < 1 {
		return fmt.Errorf("no loop dialer hosts in model")
	}

	var allHosts []*model.Host
	copy(allHosts, dialerHosts)
	allHosts = append(allHosts, listenerHosts...)

	if err := self.killPrevious(m, allHosts); err != nil {
		return fmt.Errorf("error killing previous loop hosts (%w)", err)
	}
	if err := self.startListeners(m, listenerHosts); err != nil {
		return fmt.Errorf("error starting loop listeners (%w)", err)
	}
	if err := self.startDialers(m, initiatorHost, dialerHosts, dialerIds); err != nil {
		return fmt.Errorf("error starting loop dialers (%w)", err)
	}

	return nil
}

func (_ *loopOperation) killPrevious(m *model.Model, hosts []*model.Host) error {
	for _, host := range hosts {
		ssh := fablib.NewSshConfigFactoryImpl(m, host.PublicIp)
		if err := fablib.RemoteKill(ssh, "ziti-fabric-test loop2"); err != nil {
			return fmt.Errorf("error cleaning up old loop executions (%w)", err)
		}
	}
	return nil
}

func (_ *loopOperation) startListeners(m *model.Model, listenerHosts []*model.Host) error {
	for _, listenerHost := range listenerHosts {
		ssh := fablib.NewSshConfigFactoryImpl(m, listenerHost.PublicIp)
		listenerCmd := fmt.Sprintf("nohup /home/%s/fablab/bin/ziti-fabric-test loop2 listener -b tcp:0.0.0.0:8171 >> /home/%s/ziti-fabric-test-loop2-listener.log 2>&1 &", ssh.User(), ssh.User())
		if output, err := fablib.RemoteExec(ssh, listenerCmd); err != nil {
			return fmt.Errorf("error starting loop listener [%s] (%w)", output, err)
		}
		logrus.Infof(listenerHost.PublicIp, listenerCmd)
	}
	return nil
}

func (self *loopOperation) startDialers(m *model.Model, initiatorHost, dialerHosts []*model.Host, dialerIds []string) error {
	endpoint := fmt.Sprintf("tls:%s:7002", initiatorHost[0].PublicIp)
	for i := 0; i < len(dialerHosts); i++ {
		ssh := fablib.NewSshConfigFactoryImpl(m, dialerHosts[i].PublicIp)
		dialerCmd := fmt.Sprintf("nohup /home/%s/fablab/bin/ziti-fabric-test loop2 dialer /home/%s/fablab/cfg/%s -e %s -s %s >> /home/%s/ziti-fabric-test-loop2-dialer.log 2>&1 &", ssh.User(), ssh.User(), self.loopScenario(m), endpoint, dialerIds[i], ssh.User())
		if output, err := fablib.RemoteExec(ssh, dialerCmd); err != nil {
			return fmt.Errorf("error starting loop dialer [%s] (%w)", output, err)
		}
		logrus.Infof(dialerHosts[i].PublicIp, dialerCmd)
	}
	return nil
}

func (_ *loopOperation) loopScenario(m *model.Model) string {
	return "10-ambient.loop2.yml"
}

type loopOperation struct{}