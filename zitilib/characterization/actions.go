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
	"github.com/netfoundry/fablab/kernel/model"
	zitilib_characterization_actions "github.com/netfoundry/fablab/zitilib/characterization/actions"
)

func newActionsFactory() model.Factory {
	return &actionsFactory{}
}

func (f *actionsFactory) Build(m *model.Model) error {
	m.Actions = model.ActionBinders{
		"bootstrap": zitilib_characterization_actions.NewBootstrapAction(),
		"start":     zitilib_characterization_actions.NewStartAction(),
		"stop":      zitilib_characterization_actions.NewStopAction(),
		"console":   zitilib_characterization_actions.NewConsoleAction(),
		"report":    zitilib_characterization_actions.NewReportAction(),
		"logs":      zitilib_characterization_actions.NewLogsAction(),
	}
	return nil
}

type actionsFactory struct{}
