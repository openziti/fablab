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

package zitilib_transwarp_actions

import (
	"github.com/openziti/fablab/kernel/model"
	zitilib_actions "github.com/openziti/fablab/zitilib/actions"
)

type actionsFactory struct{}

func NewActionsFactory() model.Factory {
	return &actionsFactory{}
}

func (_ *actionsFactory) Build(m *model.Model) error {
	m.Actions = model.ActionBinders{
		"bootstrap": newBootstrapAction(),
		"logs": func(_ *model.Model) model.Action { return zitilib_actions.Logs() },
	}
	return nil
}