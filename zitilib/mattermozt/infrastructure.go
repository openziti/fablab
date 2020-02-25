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

package mattermozt

import (
	"fmt"
	semaphore0 "github.com/netfoundry/fablab/kernel/fablib/runlevel/0_infrastructure/semaphore"
	terraform0 "github.com/netfoundry/fablab/kernel/fablib/runlevel/0_infrastructure/terraform"
	terraform6 "github.com/netfoundry/fablab/kernel/fablib/runlevel/6_disposal/terraform"
	"github.com/netfoundry/fablab/kernel/model"
	"time"
)

func newInfrastructureFactory() model.Factory {
	return &infrastructureFactory{}
}

func (self *infrastructureFactory) Build(m *model.Model) error {
	if err := self.buildInfrastructure(m); err != nil {
		return fmt.Errorf("error building infrastructure bindings (%w)", err)
	}
	if err := self.buildDisposal(m); err != nil {
		return fmt.Errorf("error building disposal bindings (%w)", err)
	}
	return nil
}

func (self *infrastructureFactory) buildInfrastructure(m *model.Model) error {
	m.Infrastructure = model.InfrastructureBinders{
		func(m *model.Model) model.InfrastructureStage { return terraform0.Express() },
		func(m *model.Model) model.InfrastructureStage { return semaphore0.Restart(90 * time.Second) },
	}
	return nil
}

func (self *infrastructureFactory) buildDisposal(m *model.Model) error {
	m.Disposal = model.DisposalBinders{
		func(m *model.Model) model.DisposalStage { return terraform6.Dispose() },
	}
	return nil
}

type infrastructureFactory struct{}
