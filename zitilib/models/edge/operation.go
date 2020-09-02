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

package edge

import (
	"fmt"
	fablib_5_operation "github.com/openziti/fablab/kernel/fablib/runlevel/5_operation"
	"github.com/openziti/fablab/kernel/model"
	zitilib_5_operation "github.com/openziti/fablab/zitilib/runlevel/5_operation"
	"time"
)

func newOperationFactory() model.Factory {
	return &operationFactory{}
}

// operationFactory is a model.Factory that is responsible for building and connecting the model.OperatingBinders that
// represent the operational phase of the model.
//
// In our case, this model launches mesh structure polling, fabric metrics listening, and then creates the correct
// loop2 dialers and listeners that run against the model. When the dialers complete their operation, the joiner will
// join with them, invoking the closer and ending the mesh and metrics pollers. Finally, the instance state is
// persisted as a dump.
//
func (self *operationFactory) Build(m *model.Model) error {
	closer := make(chan struct{})
	var joiners []chan struct{}

	m.Operation = append(m.Operation, model.OperatingBinders{
		func(m *model.Model) model.OperatingStage { return zitilib_5_operation.Mesh(closer) },
		func(m *model.Model) model.OperatingStage { return zitilib_5_operation.Metrics(closer) },
	}...)

	listeners, err := self.listeners(m)
	if err != nil {
		return fmt.Errorf("error creating listeners (%w)", err)
	}
	m.Operation = append(m.Operation, listeners...)

	m.Operation = append(m.Operation, func(m *model.Model) model.OperatingStage {
		return fablib_5_operation.Timer(5*time.Second, nil)
	})

	dialers, dialerJoiners, err := self.dialers(m)
	if err != nil {
		return fmt.Errorf("error creating dialers (%w)", err)
	}
	joiners = append(joiners, dialerJoiners...)
	m.Operation = append(m.Operation, dialers...)

	m.Operation = append(m.Operation, model.OperatingBinders{
		func(m *model.Model) model.OperatingStage { return fablib_5_operation.Joiner(joiners) },
		func(m *model.Model) model.OperatingStage { return fablib_5_operation.Closer(closer) },
		func(m *model.Model) model.OperatingStage { return fablib_5_operation.Persist() },
	}...)

	return nil
}

func (_ *operationFactory) listeners(m *model.Model) (binders []model.OperatingBinder, err error) {
	hosts := m.SelectHosts("@server")
	if len(hosts) < 1 {
		return nil, fmt.Errorf("no '@server' hosts in model")
	}

	for _, host := range hosts {
		boundHost := host
		binders = append(binders, func(m *model.Model) model.OperatingStage {
			return zitilib_5_operation.LoopListener(boundHost, nil)
		})
	}

	return binders, nil
}

func (_ *operationFactory) dialers(m *model.Model) (binders []model.OperatingBinder, joiners []chan struct{}, err error) {
	initiators := m.SelectHosts("@client")
	if len(initiators) != 1 {
		return nil, nil, fmt.Errorf("expected 1 '@client' host in model")
	}

	var hosts []*model.Host
	var ids []string
	for id, host := range m.MustSelectRegion("initiator").Hosts {
		if host.HasTag("server") {
			hosts = append(hosts, host)
			ids = append(ids, id)
		}
	}
	if len(hosts) < 1 {
		return nil, nil, fmt.Errorf("no '@initiator/@loop-dialer' hosts in model")
	}

	endpoint := fmt.Sprintf("tls:%s:7002", initiators[0].PublicIp)

	binders = make([]model.OperatingBinder, 0)
	for i := 0; i < len(hosts); i++ {
		joiner := make(chan struct{}, 1)
		binderHost := hosts[i]
		binderId := ids[i]
		binders = append(binders, func(m *model.Model) model.OperatingStage {
			return zitilib_5_operation.LoopDialer(binderHost, binderId, "10-ambient.loop2.yml", endpoint, joiner)
		})
		joiners = append(joiners, joiner)
	}

	return binders, joiners, nil
}

type operationFactory struct{}
