/*
	Copyright 2019 NetFoundry, Inc.

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

package operation

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/model"
)

func Actions(actions ...string) model.OperatingStage {
	return &actionOperation{actions: actions}
}

func (actionOperation *actionOperation) Operate(m *model.Model, _ string) error {
	for _, actionName := range actionOperation.actions {
		action, found := m.GetAction(actionName)
		if !found {
			return fmt.Errorf("no [%s] action", actionName)
		}
		fablib.FigletMini("action: " + actionName)
		if err := action.Execute(m); err != nil {
			return fmt.Errorf("error executing [%s] action (%w)", actionName, err)
		}
	}
	return nil
}

type actionOperation struct {
	actions []string
}
