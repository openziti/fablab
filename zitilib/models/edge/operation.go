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

	m.AddOperatingActions("syncModelEdgeState")
	m.AddOperatingStage(zitilib_5_operation.Mesh(closer))
	m.AddOperatingStage(zitilib_5_operation.MetricsWithIdMapper(closer, func(id string) string {
		return "component.edgeId:" + id
	}))

	if err := self.listeners(m); err != nil {
		return fmt.Errorf("error creating listeners (%w)", err)
	}

	m.AddOperatingStage(fablib_5_operation.Timer(5*time.Second, nil))

	if err := self.dialers(m); err != nil {
		return fmt.Errorf("error creating dialers (%w)", err)
	}

	m.AddOperatingStage(fablib_5_operation.Closer(closer))
	m.AddOperatingStage(fablib_5_operation.Persist())

	return nil
}

func (_ *operationFactory) listeners(m *model.Model) error {
	components := m.SelectComponents(models.ServiceTag)
	if len(components) < 1 {
		return fmt.Errorf("no '%v' components in model", models.ServiceTag)
	}

	for _, c := range components {
		remoteConfigFile := "/home/fedora/fablab/cfg/" + c.PublicIdentity + ".json"
		stage := zitilib_5_operation.LoopListener(c.GetHost(), nil, "edge:perf-test", "--config-file", remoteConfigFile)
		m.AddOperatingStage(stage)
	}

	return nil
}

func (_ *operationFactory) dialers(m *model.Model) error {
	components := m.SelectComponents(models.ClientTag)
	if len(components) < 1 {
		return fmt.Errorf("no '%v' components in model", models.ClientTag)
	}

	var joiners []chan struct{}

	for _, c := range components {
		joiner := make(chan struct{}, 1)
		remoteConfigFile := "/home/fedora/fablab/cfg/" + c.PublicIdentity + ".json"
		stage := zitilib_5_operation.LoopDialer(c.GetHost(), "10-ambient.loop2.yml", "edge:perf-test", joiner, "--config-file", remoteConfigFile)
		m.AddOperatingStage(stage)
		joiners = append(joiners, joiner)
	}

	m.AddOperatingStage(fablib_5_operation.Joiner(joiners))

	return nil
}

type operationFactory struct{}
