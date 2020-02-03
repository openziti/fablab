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

package zitilib_characterization_internet

import (
	"fmt"
	linked_0 "github.com/netfoundry/fablab/kernel/fablib/runlevel/0_infrastructure/linked"
	"github.com/netfoundry/fablab/kernel/model"
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
	m.Disposal = nil
	return nil
}

func (f *bindingsFactory) replaceInfrastructure(m *model.Model) error {
	m.Infrastructure = model.InfrastructureBinders{
		func(m *model.Model) model.InfrastructureStage { return linked_0.Linked() },
	}
	return nil
}

type bindingsFactory struct{}
