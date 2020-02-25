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

package mattermozt

import (
	"github.com/netfoundry/fablab/kernel/fablib/runlevel/2_kitting/devkit"
	"github.com/netfoundry/fablab/kernel/model"
	zitilib_bootstrap "github.com/netfoundry/fablab/zitilib/development/bootstrap"
)

func newKittingFactory() model.Factory {
	return &kittingFactory{}
}

func (self *kittingFactory) Build(m *model.Model) error {
	m.Kitting = model.KittingBinders{
		func(m *model.Model) model.KittingStage {
			zitiBinaries := []string{
				"ziti-controller",
				"ziti-fabric",
				"ziti-fabric-test",
				"ziti-router",
			}
			return devkit.DevKit(zitilib_bootstrap.ZitiDistBinaries(), zitiBinaries)
		},
	}
	return nil
}

type kittingFactory struct{}