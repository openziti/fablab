/*
	Copyright NetFoundry, Inc.

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

package zitilib_transwarp_actions

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib/actions"
	"github.com/openziti/fablab/kernel/fablib/actions/component"
	"github.com/openziti/fablab/kernel/fablib/actions/host"
	"github.com/openziti/fablab/kernel/fablib/actions/semaphore"
	"github.com/openziti/fablab/kernel/model"
	actions2 "github.com/openziti/fablab/zitilib/actions"
	"github.com/openziti/fablab/zitilib/models"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

type bootstrapAction struct{}

func newBootstrapAction() model.ActionBinder {
	action := &bootstrapAction{}
	return action.bind
}

func (_ *bootstrapAction) bind(m *model.Model) model.Action {
	workflow := actions.Workflow()

	/*
	 * Restart controller with new database.
	 */
	workflow.AddAction(component.Stop(models.ControllerTag))
	workflow.AddAction(host.Exec(m.MustSelectHost(models.HasControllerComponent), "rm -f ~/ctrl.db"))
	workflow.AddAction(component.Start(models.ControllerTag))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))

	/*
	 * Create routers.
	 */
	for _, router := range m.SelectComponents(models.RouterTag) {
		certPath := filepath.Join(model.PkiBuild(), fmt.Sprintf("/intermediate/certs/%s-client.cert", router.PublicIdentity))
		workflow.AddAction(actions2.Fabric("create", "router", certPath))
	}

	/*
	 * Create services and terminators.
	 */
	iperfServer := m.MustSelectHost(".iperf_server")
	terminatingRouters := m.SelectComponents(models.RemoteId)

	if len(terminatingRouters) != 1 {
		logrus.Fatalf("expect 1 terminating router, got [%d]", len(terminatingRouters))
	}
	workflow.AddAction(actions2.Fabric("create", "service", "iperf"))
	workflow.AddAction(actions2.Fabric("create", "terminator", "iperf", terminatingRouters[0].PublicIdentity, "tcp:"+iperfServer.PrivateIp+":7001"))

	workflow.AddAction(actions2.Fabric("create", "service", "ssh"))
	workflow.AddAction(actions2.Fabric("create", "terminator", "ssh", terminatingRouters[0].PublicIdentity, "tcp:127.0.0.1:22"))
	workflow.AddAction(actions2.Fabric("create", "service", "nc4506"))
	workflow.AddAction(actions2.Fabric("create", "terminator", "nc4506", terminatingRouters[0].PublicIdentity, "tcp:127.0.0.1:4506"))

	/*
	 * Stop controller.
	 */
	workflow.AddAction(component.Stop(models.ControllerTag))

	return workflow
}
