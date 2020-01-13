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

package zitilab_characterization_internet

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	linked_0 "github.com/netfoundry/fablab/kernel/runlevel/0_infrastructure/linked"
	operation "github.com/netfoundry/fablab/kernel/runlevel/5_operation"
	"time"
)

func newBindingsFactory() model.Factory {
	return &bindingsFactory{}
}

func (f *bindingsFactory) Build(m *model.Model) error {
	m.Actions = nil
	if err := f.replaceInfrastructure(m); err != nil {
		return fmt.Errorf("error building infrastructure bindings (%w)", err)
	}
	m.Configuration = nil
	m.Kitting = nil
	m.Distribution = nil
	m.Activation = nil
	if err := f.replaceOperation(m); err != nil {
		return fmt.Errorf("error building operation bindings (%w)", err)
	}
	m.Disposal = nil
	return nil
}

func (f *bindingsFactory) replaceInfrastructure(m *model.Model) error {
	m.Infrastructure = model.InfrastructureBinders{
		func(m *model.Model) model.InfrastructureStage { return linked_0.Linked() },
	}
	return nil
}

func (f *bindingsFactory) replaceOperation(m *model.Model) error {
	values := m.GetHosts("local", "service")
	var directEndpoint string
	if len(values) == 1 {
		directEndpoint = values[0].PublicIp
	} else {
		return fmt.Errorf("need single host for local:@service, found [%d]", len(values))
	}

	values = m.GetHosts("short", "initiator")
	var shortProxy string
	if len(values) == 1 {
		shortProxy = values[0].PublicIp
	} else {
		return fmt.Errorf("need single host for short:initiator, found [%d]", len(values))
	}

	values = m.GetHosts("medium", "initiator")
	var mediumProxy string
	if len(values) == 1 {
		mediumProxy = values[0].PublicIp
	} else {
		return fmt.Errorf("need a single host for medium:initiator, found [%d]", len(values))
	}

	values = m.GetHosts("long", "initiator")
	var longProxy string
	if len(values) == 1 {
		longProxy = values[0].PublicIp
	} else {
		return fmt.Errorf("need a single host for long:initiator, found [%d]", len(values))
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

		func(m *model.Model) model.OperatingStage { return operation.Iperf(directEndpoint, "local", "service", "short", "client", int(sampleDuration.Seconds())) },
		func(m *model.Model) model.OperatingStage { return operation.Iperf(shortProxy, "local", "service", "short", "client", int(sampleDuration.Seconds())) },

		func(m *model.Model) model.OperatingStage { return operation.Iperf(directEndpoint, "local", "service", "medium", "client", int(sampleDuration.Seconds())) },
		func(m *model.Model) model.OperatingStage { return operation.Iperf(mediumProxy, "local", "service", "medium", "client", int(sampleDuration.Seconds())) },

		func(m *model.Model) model.OperatingStage { return operation.Iperf(directEndpoint, "local", "service", "long", "client", int(sampleDuration.Seconds())) },
		func(m *model.Model) model.OperatingStage { return operation.Iperf(longProxy, "local", "service", "long", "client", int(sampleDuration.Seconds())) },

		func(m *model.Model) model.OperatingStage { return operation.Closer(c) },
		func(m *model.Model) model.OperatingStage { return operation.Persist() },
	}

	return nil
}

type bindingsFactory struct{}
