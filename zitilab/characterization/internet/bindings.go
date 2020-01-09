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

package zitilab_characterization_internet

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	linked_0 "github.com/netfoundry/fablab/kernel/runlevel/0_infrastructure/linked"
)

func newBindingsFactory() model.Factory {
	return &bindingsFactory{}
}

func (f *bindingsFactory) Build(m *model.Model) error {
	if err := f.buildInfrastructure(m); err != nil {
		return fmt.Errorf("error building infrastructure bindings (%w)", err)
	}
	if err := f.buildDisposal(m); err != nil {
		return fmt.Errorf("error building disposal bindings (%w)", err)
	}
	return nil
}

func (f *bindingsFactory) buildInfrastructure(m *model.Model) error {
	m.Infrastructure = model.InfrastructureBinders{
		func(m *model.Model) model.InfrastructureStage { return linked_0.Linked() },
	}
	return nil
}

func (f *bindingsFactory) buildDisposal(m *model.Model) error {
	m.Disposal = nil
	return nil
}

type bindingsFactory struct{}
