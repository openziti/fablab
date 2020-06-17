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

package zitilib_characterization_actions

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib/actions"
	"github.com/openziti/fablab/kernel/fablib/actions/component"
	"github.com/openziti/fablab/kernel/fablib/actions/host"
	"github.com/openziti/fablab/kernel/fablib/actions/semaphore"
	"github.com/openziti/fablab/kernel/model"
	actions2 "github.com/openziti/fablab/zitilib/actions"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

func NewBootstrapAction() model.ActionBinder {
	action := &bootstrapAction{}
	return action.bind
}

func (a *bootstrapAction) bind(m *model.Model) model.Action {
	sshUsername := m.Variables.Must("credentials", "ssh", "username").(string)

	workflow := actions.Workflow()

	workflow.AddAction(component.Stop("@ctrl", "@ctrl", "@ctrl"))
	workflow.AddAction(host.Exec(m.GetHostByTags("ctrl", "ctrl"), "rm -f ~/ctrl.db"))
	workflow.AddAction(component.Start("@ctrl", "@ctrl", "@ctrl"))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))

	for _, router := range m.GetComponentsByTag("router") {
		cert := fmt.Sprintf("/intermediate/certs/%s-client.cert", router.PublicIdentity)
		workflow.AddAction(actions2.Fabric("create", "router", filepath.Join(model.PkiBuild(), cert)))
	}

	iperfServer := m.GetHostByTags("iperf_server", "iperf_server")
	if iperfServer != nil {
		terminatingRouters := m.GetComponentsByTag("terminator")
		if len(terminatingRouters) < 1 {
			logrus.Fatal("need at least 1 terminating router!")
		}
		workflow.AddAction(actions2.Fabric("create", "service", "iperf"))
		workflow.AddAction(actions2.Fabric("create", "terminator", "iperf", terminatingRouters[0].PublicIdentity, "tcp:"+iperfServer.PublicIp+":7001"))
		workflow.AddAction(actions2.Fabric("create", "service", "iperf_udp"))
		workflow.AddAction(actions2.Fabric("create", "terminator", "iperf_udp", terminatingRouters[0].PublicIdentity, "udp:"+iperfServer.PublicIp+":7001", "--binding", "transport_udp"))
	}

	for _, h := range m.GetAllHosts() {
		workflow.AddAction(host.Exec(h, fmt.Sprintf("mkdir -p /home/%s/.ziti", sshUsername)))
		workflow.AddAction(host.Exec(h, fmt.Sprintf("rm -f /home/%s/.ziti/identities.yml", sshUsername)))
		workflow.AddAction(host.Exec(h, fmt.Sprintf("ln -s /home/%s/fablab/cfg/remote_identities.yml /home/%s/.ziti/identities.yml", sshUsername, sshUsername)))
	}

	workflow.AddAction(component.Stop("@ctrl", "@ctrl", "@ctrl"))

	return workflow
}

type bootstrapAction struct{}
