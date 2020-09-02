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
	"github.com/openziti/fablab/kernel/fablib/actions"
	"github.com/openziti/fablab/kernel/fablib/actions/component"
	"github.com/openziti/fablab/kernel/fablib/actions/semaphore"
	"github.com/openziti/fablab/kernel/model"
	"time"
)

type startAction struct{}

func newStartAction() model.ActionBinder {
	action := &startAction{}
	return action.bind
}
func (_ *startAction) bind(_ *model.Model) model.Action {
	return actions.Workflow(
		component.Start("@ctrl"),
		semaphore.Sleep(1*time.Second),
		component.Start("@router"),
		semaphore.Sleep(1*time.Second),
	)
}
