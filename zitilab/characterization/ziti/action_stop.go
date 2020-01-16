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
	"github.com/netfoundry/fablab/kernel/actions"
	"github.com/netfoundry/fablab/kernel/actions/component"
	"github.com/netfoundry/fablab/kernel/model"
)

func newStopAction() model.ActionBinder {
	action := &stopAction{}
	return action.bind
}

func (a *stopAction) bind(m *model.Model) model.Action {
	return actions.Workflow(
		component.Stop("@router", "@router", "@router"),
		component.Stop("@ctrl", "@ctrl", "@ctrl"),
	)
}

type stopAction struct{}
