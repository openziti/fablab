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

package zitilib_examples_actions

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib/actions"
	"github.com/netfoundry/fablab/kernel/fablib/actions/component"
	"github.com/netfoundry/fablab/kernel/fablib/actions/host"
	"github.com/netfoundry/fablab/kernel/fablib/actions/semaphore"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"time"
)

func NewStartAction() model.ActionBinder {
	action := startAction{}
	return action.bind
}

func (self *startAction) bind(m *model.Model) model.Action {
	sshUsername := m.MustVariable("credentials", "ssh", "username").(string)

	listenerCmd := fmt.Sprintf("nohup /home/%s/fablab/bin/ziti-fabric-test loop2 listener -b tcp:0.0.0.0:8171 > /home/%s/ziti-fabric-test.log 2>&1 &", sshUsername, sshUsername)

	workflow := actions.Workflow()
	workflow.AddAction(component.Start("@ctrl", "@ctrl", "@ctrl"))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))
	workflow.AddAction(component.Start("@router", "@router", "@router"))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))
	workflow.AddAction(host.GroupExec("@loop", "@loop-listener", listenerCmd))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))

	r001 := m.GetHosts("@initiator", "@initiator")
	if len(r001) != 1 {
		logrus.Fatalf("expected to find a single host tagged [initiator/initiator]")
	}
	endpoint := fmt.Sprintf("tls:%s:7001", r001[0].PublicIp)
	dialerActions, err := self.createDialerActions(m, endpoint)
	if err != nil {
		logrus.Fatalf("error creating dialer actions (%w)", err)
	}
	for _, dialerAction := range dialerActions {
		workflow.AddAction(dialerAction)
	}

	return workflow
}

func (self *startAction) createDialerActions(m *model.Model, endpoint string) ([]model.Action, error) {
	initiatorRegion := m.GetRegionByTag("initiator")
	if initiatorRegion == nil {
		return nil, fmt.Errorf("unable to find 'initiator' region")
	}

	sshUsername := m.MustVariable("credentials", "ssh", "username").(string)
	loopScenario := self.loopScenario(m)
	dialerActions := make([]model.Action, 0)
	for hostId, h := range initiatorRegion.Hosts {
		for _, tag := range h.Tags {
			if tag == "loop-dialer" {
				dialerCmd := fmt.Sprintf("nohup /home/%s/fablab/bin/ziti-fabric-test loop2 dialer /home/%s/fablab/cfg/%s -e %s -s %s > /home/%s/ziti-fabric-test.log 2>&1 &", sshUsername, sshUsername, loopScenario, endpoint, hostId, sshUsername)
				dialerActions = append(dialerActions, host.Exec(h, dialerCmd))
			}
		}
	}

	return dialerActions, nil
}

func (_ *startAction) loopScenario(m *model.Model) string {
	loopScenario := "10-ambient.loop2.yml"
	if initiator := m.GetRegionByTag("initiator"); initiator != nil {
		if len(initiator.Hosts) > 1 {
			loopScenario = "4k-chatter.loop2.yml"
		}
	}
	return loopScenario
}

type startAction struct{}
