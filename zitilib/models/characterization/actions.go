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

package zitilib_characterization

import (
	"github.com/openziti/fablab/kernel/model"
	zitilib_actions "github.com/openziti/fablab/zitilib/actions"
	"github.com/openziti/fablab/zitilib/console"
	zitilib_characterization_actions "github.com/openziti/fablab/zitilib/models/characterization/actions"
	"github.com/openziti/fablab/zitilib/models/characterization/reporting"
)

func newActionsFactory() model.Factory {
	return &actionsFactory{}
}

func (f *actionsFactory) Build(m *model.Model) error {
	m.Actions = model.ActionBinders{
		"bootstrap": zitilib_characterization_actions.NewBootstrapAction(),
		"start":     zitilib_characterization_actions.NewStartAction(),
		"stop":      zitilib_characterization_actions.NewStopAction(),
		"report":    func(m *model.Model) model.Action { return reporting.Report() },
		"console":   func(m *model.Model) model.Action { return console.Console() },
		"logs":      func(m *model.Model) model.Action { return zitilib_actions.Logs() },
	}
	return nil
}

type actionsFactory struct{}
