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

package zitilib_examples

import (
	"fmt"
	fablib_5_operation "github.com/openziti/fablab/kernel/fablib/runlevel/5_operation"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/zitilib/models"
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
		func(*model.Model) model.OperatingStage { return zitilib_5_operation.Mesh(closer) },
		func(*model.Model) model.OperatingStage { return zitilib_5_operation.Metrics(closer) },
	}...)

	listeners, err := self.listeners(m)
	if err != nil {
		return fmt.Errorf("error creating listeners (%w)", err)
	}
	m.Operation = append(m.Operation, listeners...)

	m.Operation = append(m.Operation, func(*model.Model) model.OperatingStage {
		return fablib_5_operation.Timer(5*time.Second, nil)
	})

	dialers, dialerJoiners, err := self.dialers(m)
	if err != nil {
		return fmt.Errorf("error creating dialers (%w)", err)
	}
	joiners = append(joiners, dialerJoiners...)
	m.Operation = append(m.Operation, dialers...)

	m.Operation = append(m.Operation, model.OperatingBinders{
		func(*model.Model) model.OperatingStage { return fablib_5_operation.Joiner(joiners) },
		func(*model.Model) model.OperatingStage { return fablib_5_operation.Closer(closer) },
		func(*model.Model) model.OperatingStage { return fablib_5_operation.Persist() },
	}...)

	return nil
}

func (_ *operationFactory) listeners(m *model.Model) (binders []model.OperatingBinder, err error) {
	hosts := m.SelectHosts(models.LoopListenerTag)
	if len(hosts) < 1 {
		return nil, fmt.Errorf("no '%v' hosts in model", models.LoopListenerTag)
	}

	for _, host := range hosts {
		boundHost := host
		binders = append(binders, func(*model.Model) model.OperatingStage {
			return zitilib_5_operation.LoopListener(boundHost, nil, "tcp:0.0.0.0:8171")
		})
	}

	return binders, nil
}

func (_ *operationFactory) dialers(m *model.Model) (binders []model.OperatingBinder, joiners []chan struct{}, err error) {
	initiator, err := m.SelectHost("component.initiator.router")
	if err != nil {
		return nil, nil, err
	}

	endpoint := fmt.Sprintf("tls:%s:7002", initiator.PublicIp)

	binders = make([]model.OperatingBinder, 0)

	hosts, err := m.MustSelectHosts(models.LoopDialerTag, 1)
	if err != nil {
		return nil, nil, err
	}

	for _, host := range hosts {
		boundHost := host
		joiner := make(chan struct{}, 1)
		binders = append(binders, func(*model.Model) model.OperatingStage {
			return zitilib_5_operation.LoopDialer(boundHost, "10-ambient.loop2.yml", endpoint, joiner)
		})
		joiners = append(joiners, joiner)
	}

	return binders, joiners, nil
}

type operationFactory struct{}
