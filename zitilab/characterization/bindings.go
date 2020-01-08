/*
	Copyright 2020 Netfoundry, Inc.

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

package characterization

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	semaphore0 "github.com/netfoundry/fablab/kernel/runlevel/0_infrastructure/semaphore"
	terraform0 "github.com/netfoundry/fablab/kernel/runlevel/0_infrastructure/terraform"
	"time"
)

func newBindingsFactory() *bindingsFactory {
	return &bindingsFactory{}
}

func (f *bindingsFactory) Build(m *model.Model) error {
	if err := f.buildInfrastructure(m); err != nil {
		return fmt.Errorf("error building infrastructure bindings (%w)", err)
	}
	return nil
}

func (f *bindingsFactory) buildInfrastructure(m *model.Model) error {
	m.Infrastructure = model.InfrastructureBinders{
		func(m *model.Model) model.InfrastructureStage { return terraform0.Express() },
		func(m *model.Model) model.InfrastructureStage { return semaphore0.Restart(90 * time.Second) },
	}
	return nil
}

type bindingsFactory struct{}