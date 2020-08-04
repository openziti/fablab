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

package dilithium

import (
	"github.com/openziti/fablab/kernel/fablib/actions"
	"github.com/openziti/fablab/kernel/model"
	zitilib_actions "github.com/openziti/fablab/zitilib/actions"
	dilithium_actions "github.com/openziti/fablab/zitilib/models/development/dilithium/actions"
)

type actionsFactory struct{}

func newActionsFactory() model.Factory {
	return &actionsFactory{}
}

func (self *actionsFactory) Build(m *model.Model) error {
	m.Actions = model.ActionBinders{
		"logs": func(m *model.Model) model.Action {
			return actions.Workflow(
				zitilib_actions.Logs(),
				dilithium_actions.Clean(),
			)
		},
	}
	return nil
}