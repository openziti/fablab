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

package actions

import (
	"github.com/openziti/fablab/kernel/fablib/actions"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/zitilib/actions/edge"
	"github.com/openziti/fablab/zitilib/models"
)

func NewSyncModelEdgeStateAction() model.ActionBinder {
	action := &syncModelEdgeStateAction{}
	return action.bind
}

func (a *syncModelEdgeStateAction) bind(*model.Model) model.Action {
	workflow := actions.Workflow()
	workflow.AddAction(edge.Login(models.HasControllerComponent))
	workflow.AddAction(edge.SyncModelEdgeState(".edge-router"))
	return workflow
}

type syncModelEdgeStateAction struct{}
