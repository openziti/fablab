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

package actions

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib/actions"
	"github.com/openziti/fablab/kernel/fablib/actions/component"
	"github.com/openziti/fablab/kernel/fablib/actions/host"
	"github.com/openziti/fablab/kernel/fablib/actions/semaphore"
	"github.com/openziti/fablab/kernel/model"
	zitilib_actions "github.com/openziti/fablab/zitilib/actions"
	"github.com/openziti/fablab/zitilib/actions/edge"
	"github.com/openziti/fablab/zitilib/models"
	"path/filepath"
	"time"
)

func NewBootstrapAction() model.ActionBinder {
	action := &bootstrapAction{}
	return action.bind
}

func (a *bootstrapAction) bind(m *model.Model) model.Action {
	sshUsername := m.MustVariable("credentials", "ssh", "username").(string)

	workflow := actions.Workflow()

	workflow.AddAction(component.Stop(models.ControllerTag))
	workflow.AddAction(edge.InitController(models.ControllerTag))
	workflow.AddAction(component.Start(models.ControllerTag))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))

	for _, router := range m.SelectComponents(models.RouterTag) {
		cert := fmt.Sprintf("/intermediate/certs/%s-client.cert", router.PublicIdentity)
		workflow.AddAction(zitilib_actions.Fabric("create", "router", filepath.Join(model.PkiBuild(), cert)))
	}

	workflow.AddAction(host.GroupExec("*", true,
		fmt.Sprintf("mkdir -p /home/%s/.ziti", sshUsername),
		fmt.Sprintf("rm -f /home/%s/.ziti/identities.yml", sshUsername),
		fmt.Sprintf("ln -s /home/%s/fablab/cfg/remote_identities.yml /home/%s/.ziti/identities.yml", sshUsername, sshUsername),
	))

	workflow.AddAction(edge.Login(models.HasControllerComponent))

	workflow.AddAction(component.StopInParallel(models.EdgeRouterTag))
	workflow.AddAction(edge.InitEdgeRouters(models.EdgeRouterTag, true))
	workflow.AddAction(edge.InitIdentities(models.SdkAppTag, true))

	workflow.AddAction(zitilib_actions.Edge("create", "service", "perf-test"))
	workflow.AddAction(zitilib_actions.Edge("create", "service-policy", "perf-bind", "Bind", "--service-roles", "@perf-test", "--identity-roles", "#service"))
	workflow.AddAction(zitilib_actions.Edge("create", "service-policy", "perf-dial", "Dial", "--service-roles", "@perf-test", "--identity-roles", "#client"))
	workflow.AddAction(zitilib_actions.Edge("create", "edge-router-policy", "client-routers", "--edge-router-roles", "#initiator", "--identity-roles", "#client"))
	workflow.AddAction(zitilib_actions.Edge("create", "edge-router-policy", "server-routers", "--edge-router-roles", "#terminator", "--identity-roles", "#service"))
	workflow.AddAction(zitilib_actions.Edge("create", "service-edge-router-policy", "blanket", "--edge-router-roles", "#all", "--service-roles", "#all"))

	workflow.AddAction(component.Stop(models.ControllerTag))

	return workflow
}

type bootstrapAction struct{}
