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

package zitilib_examples

import (
	operation "github.com/netfoundry/fablab/kernel/fablib/runlevel/5_operation"
	"github.com/netfoundry/fablab/kernel/model"
	__operation "github.com/netfoundry/fablab/zitilib/runlevel/5_operation"
	"time"
)

func newOperationFactory() model.Factory {
	return &operationFactory{}
}

func (_ *operationFactory) Build(m *model.Model) error {
	c := make(chan struct{})
	binders := model.OperatingBinders{
		func(m *model.Model) model.OperatingStage { return __operation.Mesh(c) },
		func(m *model.Model) model.OperatingStage { return __operation.Metrics(c) },
		func(m *model.Model) model.OperatingStage { return operation.Timer(5 * time.Minute, c) },
		func(m *model.Model) model.OperatingStage { return operation.Persist() },
	}
	m.Operation = binders

	return nil
}

type operationFactory struct{}
