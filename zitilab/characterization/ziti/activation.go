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

package zitilab_characterization_ziti

import (
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/kernel/runlevel/4_activation/action"
)

func newActivationFactory() model.Factory {
	return &activationFactory{}
}

func (f *activationFactory) Build(m *model.Model) error {
	m.Activation = model.ActivationBinders{
		func(m *model.Model) model.ActivationStage {
			return action.Activation("bootstrap", "start")
		},
	}
	return nil
}

type activationFactory struct{}