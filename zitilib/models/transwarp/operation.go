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
	zitilib_runlevel_5_operation "github.com/openziti/fablab/zitilib/runlevel/5_operation"
)

type operationFactory struct{}

func newOperationFactory() model.Factory {
	return &operationFactory{}
}

func (_ *operationFactory) Build(m *model.Model) error {
	directEndpoint := m.MustSelectHost("local", "service").PublicIp
	remoteProxy := m.MustSelectHost("remote", "router").PrivateIp

	c := make(chan struct{})
	m.Operation = model.OperatingBinders{
		func(_ *model.Model) model.OperatingStage { return zitilib_runlevel_5_operation.Metrics(c) },

		func(_ *model.Model) model.OperatingStage { return operation.Banner("Transwarp") },
		func(_ *model.Model) model.OperatingStage { return operation.Iperf("ziti", remoteProxy, "local", "service", "remote", "client", 120) },
		func(_ *model.Model) model.OperatingStage { return operation.Persist() },

		func(_ *model.Model) model.OperatingStage { return operation.Banner("Direct Internet") },
		func(_ *model.Model) model.OperatingStage { return operation.Iperf("internet", directEndpoint, "local", "service", "remote", "client", 120)},
		func(_ *model.Model) model.OperatingStage { return operation.Persist() },

		func(_ *model.Model) model.OperatingStage { return operation.Closer(c) },
		func(_ *model.Model) model.OperatingStage { return operation.Persist() },
	}

	return nil
}