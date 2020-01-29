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

package zitilab_characterization_ziti

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	operation "github.com/netfoundry/fablab/kernel/runlevel/5_operation"
	"time"
)

func newOperationFactory() model.Factory {
	return &operationFactory{}
}

func (f *operationFactory) Build(m *model.Model) error {
	values := m.GetHosts("local", "service")
	var directEndpoint string
	if len(values) == 1 {
		directEndpoint = values[0].PublicIp
	} else {
		return fmt.Errorf("need single host for local:@service, found [%d]", len(values))
	}

	values = m.GetHosts("short", "short")
	var shortProxy string
	if len(values) == 1 {
		shortProxy = values[0].PrivateIp
	} else {
		return fmt.Errorf("need single host for short:short, found [%d]", len(values))
	}

	values = m.GetHosts("medium", "medium")
	var mediumProxy string
	if len(values) == 1 {
		mediumProxy = values[0].PrivateIp
	} else {
		return fmt.Errorf("need a single host for medium:medium, found [%d]", len(values))
	}

	values = m.GetHosts("long", "long")
	var longProxy string
	if len(values) == 1 {
		longProxy = values[0].PrivateIp
	} else {
		return fmt.Errorf("need a single host for long:long, found [%d]", len(values))
	}

	minutes, found := m.GetVariable("sample_minutes")
	if !found {
		minutes = 1
	}
	sampleDuration := time.Duration(minutes.(int)) * time.Minute

	c := make(chan struct{})
	m.Operation = model.OperatingBinders{
		func(m *model.Model) model.OperatingStage { return operation.Mesh(c) },
		func(m *model.Model) model.OperatingStage { return operation.Metrics(c) },

		/*
		 * short -> local
		 */
		func(m *model.Model) model.OperatingStage {
			return operation.Iperf("ziti", shortProxy, "local", "service", "short", "client", int(sampleDuration.Seconds()))
		},
		func(m *model.Model) model.OperatingStage {
			return operation.Iperf("internet", directEndpoint, "local", "service", "short", "client", int(sampleDuration.Seconds()))
		},
		func(m *model.Model) model.OperatingStage {
			return operation.IperfUdp("ziti_1m", shortProxy, "local", "service", "short", "client", "1M", int(sampleDuration.Seconds()))
		},
		func(m *model.Model) model.OperatingStage {
			return operation.IperfUdp("internet_1m", directEndpoint, "local", "service", "short", "client", "1M", int(sampleDuration.Seconds()))
		},
		/* */

		/*
		 * medium -> local
		 */
		func(m *model.Model) model.OperatingStage {
			return operation.Iperf("ziti", mediumProxy, "local", "service", "medium", "client", int(sampleDuration.Seconds()))
		},
		func(m *model.Model) model.OperatingStage {
			return operation.Iperf("internet", directEndpoint, "local", "service", "medium", "client", int(sampleDuration.Seconds()))
		},
		func(m *model.Model) model.OperatingStage {
			return operation.IperfUdp("ziti_1m", mediumProxy, "local", "service", "medium", "client", "1M", int(sampleDuration.Seconds()))
		},
		func(m *model.Model) model.OperatingStage {
			return operation.IperfUdp("internet_1m", directEndpoint, "local", "service", "medium", "client", "1M", int(sampleDuration.Seconds()))
		},
		/* */

		/*
		 * long -> local
		 */
		func(m *model.Model) model.OperatingStage {
			return operation.Iperf("ziti", longProxy, "local", "service", "long", "client", int(sampleDuration.Seconds()))
		},
		func(m *model.Model) model.OperatingStage {
			return operation.Iperf("internet", directEndpoint, "local", "service", "long", "client", int(sampleDuration.Seconds()))
		},
		func(m *model.Model) model.OperatingStage {
			return operation.IperfUdp("ziti_1m", longProxy, "local", "service", "long", "client", "1M", int(sampleDuration.Seconds()))
		},
		func(m *model.Model) model.OperatingStage {
			return operation.IperfUdp("internet_1m", directEndpoint, "local", "service", "long", "client", "1M", int(sampleDuration.Seconds()))
		},
		/* */

		func(m *model.Model) model.OperatingStage { return operation.Closer(c) },
		func(m *model.Model) model.OperatingStage { return operation.Dumper() },
		func(m *model.Model) model.OperatingStage { return operation.Persist() },
	}

	return nil
}

type operationFactory struct{}
