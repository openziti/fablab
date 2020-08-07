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
	semaphore0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/semaphore"
	terraform0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/terraform"
	terraform6 "github.com/openziti/fablab/kernel/fablib/runlevel/6_disposal/terraform"
	"github.com/openziti/fablab/kernel/model"
	"time"
)

type infrastructureFactory struct{}

func newInfrastructureFactory() model.Factory {
	return &infrastructureFactory{}
}

func (self *infrastructureFactory) Build(m *model.Model) error {
	self.buildInfrastructure(m)
	self.buildDisposal(m)
	return nil
}

func (_ *infrastructureFactory) buildInfrastructure(m *model.Model) {
	m.Infrastructure = model.InfrastructureBinders{
		func(_ *model.Model) model.InfrastructureStage { return terraform0.Express() },
		func(_ *model.Model) model.InfrastructureStage { return semaphore0.Restart(90 * time.Second) },
	}
}

func (_ *infrastructureFactory) buildDisposal(m *model.Model) {
	m.Disposal = model.DisposalBinders{
		func(_ *model.Model) model.DisposalStage { return terraform6.Dispose() },
	}
}
