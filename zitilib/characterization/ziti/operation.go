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

package zitilib_characterization_ziti

import (
	"fmt"
	operation "github.com/netfoundry/fablab/kernel/fablib/runlevel/5_operation"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/zitilib/characterization/runlevel/5_operation"
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

	minutes := m.MustVariable("characterization", "sample_minutes")
	seconds := int((time.Duration(minutes.(int)) * time.Minute).Seconds())

	c := make(chan struct{})
	m.Operation = model.OperatingBinders{
		func(m *model.Model) model.OperatingStage { return __operation.Mesh(c) },
		func(m *model.Model) model.OperatingStage { return __operation.Metrics(c) },
	}

	m.Operation = append(m.Operation, f.sarHosts(c, m, 1, 5)...)

	m.Operation = append(m.Operation, f.forRegion("short", shortProxy, directEndpoint, seconds)...)
	m.Operation = append(m.Operation, f.forRegion("medium", mediumProxy, directEndpoint, seconds)...)
	m.Operation = append(m.Operation, f.forRegion("long", longProxy, directEndpoint, seconds)...)

	m.Operation = append(m.Operation, []model.OperatingBinder{
		func(m *model.Model) model.OperatingStage { return operation.Closer(c) },
		func(m *model.Model) model.OperatingStage { return operation.Persist() },
	}...)

	return nil
}

func (f *operationFactory) sarHosts(closer chan struct{}, m *model.Model, intervalSeconds, snapshots int) []model.OperatingBinder {
	binders := make([]model.OperatingBinder, 0)
	for _, host := range m.GetAllHosts() {
		h := host
		stage := func(m *model.Model) model.OperatingStage {
			return operation.Sar(closer, h, intervalSeconds, snapshots)
		}
		binders = append(binders, stage)
	}
	return binders
}

func (f *operationFactory) forRegion(region, initiatingRouter, directEndpoint string, seconds int) []model.OperatingBinder {
	return []model.OperatingBinder{
		func(m *model.Model) model.OperatingStage {
			return operation.Tcpdump("ziti", region, "client", 128)
		},
		func(m *model.Model) model.OperatingStage {
			return operation.Iperf("ziti", initiatingRouter, "local", "service", region, "client", seconds)
		},
		func(m *model.Model) model.OperatingStage {
			return operation.TcpdumpCloser(region, "client")
		},
		func(m *model.Model) model.OperatingStage {
			return operation.Persist()
		},

		func(m *model.Model) model.OperatingStage {
			return operation.Tcpdump("internet", region, "client", 128)
		},
		func(m *model.Model) model.OperatingStage {
			return operation.Iperf("internet", directEndpoint, "local", "service", region, "client", seconds)
		},
		func(m *model.Model) model.OperatingStage {
			return operation.TcpdumpCloser(region, "client")
		},
		func(m *model.Model) model.OperatingStage {
			return operation.Persist()
		},

		func(m *model.Model) model.OperatingStage {
			return operation.Retrieve(region, "client", ".", ".pcap")
		},

		func(m *model.Model) model.OperatingStage {
			return operation.IperfUdp("ziti_1m", initiatingRouter, "local", "service", region, "client", "1M", seconds)
		},
		func(m *model.Model) model.OperatingStage {
			return operation.IperfUdp("internet_1m", directEndpoint, "local", "service", region, "client", "1M", seconds)
		},
		func(m *model.Model) model.OperatingStage {
			return operation.Persist()
		},
	}
}

type operationFactory struct{}
