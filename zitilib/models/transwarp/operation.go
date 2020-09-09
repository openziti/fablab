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

package transwarp

import (
	operation "github.com/openziti/fablab/kernel/fablib/runlevel/5_operation"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/zitilib/models"
	zitilib_runlevel_5_operation "github.com/openziti/fablab/zitilib/runlevel/5_operation"
)

type operationFactory struct{}

func newOperationFactory() model.Factory {
	return &operationFactory{}
}

func (_ *operationFactory) Build(m *model.Model) error {
	directEndpoint := m.MustSelectHost(models.RemoteId).PublicIp
	remoteProxy := m.MustSelectHost(models.LocalId).PrivateIp

	c := make(chan struct{})
	m.Operation = model.OperatingStages{
		zitilib_runlevel_5_operation.Metrics(c),

		operation.Banner("transwarp"),
		operation.Iperf("ziti", remoteProxy, models.RemoteId, models.LocalId, 30),
		operation.Persist(),

		operation.Banner("internet"),
		operation.Iperf("internet", directEndpoint, models.RemoteId, models.LocalId, 30),
		operation.Persist(),

		operation.Closer(c),
		operation.Persist(),
	}

	return nil
}
