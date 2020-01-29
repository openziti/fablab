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

package operation

import "github.com/netfoundry/fablab/kernel/model"

func Dumper() model.OperatingStage {
	return &dumper{}
}

func (dumper *dumper) Operate(m *model.Model) error {
	if m.Data == nil {
		m.Data = make(model.Data)
	}
	m.Data["dump"] = m.Dump()
	return nil
}

type dumper struct {}