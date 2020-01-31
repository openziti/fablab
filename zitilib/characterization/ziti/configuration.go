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

package zitilib_characterization_ziti

import (
	"github.com/netfoundry/fablab/kernel/fablib/runlevel/1_configuration/config"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/zitilib/characterization/runlevel/1_configuration/pki"
)

func newConfigurationFactory() model.Factory {
	return &configurationFactory{}
}

func (f *configurationFactory) Build(m *model.Model) error {
	m.Configuration = model.ConfigurationBinders{
		func(m *model.Model) model.ConfigurationStage { return pki.Group(pki.Fabric(), pki.DotZiti()) },
		func(m *model.Model) model.ConfigurationStage { return config.Component() },
	}
	return nil
}

type configurationFactory struct{}
