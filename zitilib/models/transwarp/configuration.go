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

package transwarp

import (
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/fablib/runlevel/1_configuration/config"
	"github.com/openziti/fablab/kernel/fablib/runlevel/2_kitting/devkit"
	"github.com/openziti/fablab/kernel/model"
	zitilib_bootstrap "github.com/openziti/fablab/zitilib"
	zitilib_runlevel_1_configuration "github.com/openziti/fablab/zitilib/runlevel/1_configuration"
	"github.com/pkg/errors"
	"path/filepath"
)

type configurationFactory struct{}

func newConfigurationFactory() model.Factory {
	return &configurationFactory{}
}

func (_ *configurationFactory) Build(m *model.Model) error {
	m.Configuration = model.ConfigurationBinders{
		func(*model.Model) model.ConfigurationStage {
			return zitilib_runlevel_1_configuration.IfNoPki(zitilib_runlevel_1_configuration.Fabric(), zitilib_runlevel_1_configuration.DotZiti())
		},
		func(*model.Model) model.ConfigurationStage { return config.Component() },
		func(*model.Model) model.ConfigurationStage {
			return &kit{}
		},
		func(*model.Model) model.ConfigurationStage {
			return devkit.DevKit(zitilib_bootstrap.ZitiDistBinaries(), []string{"ziti-controller", "ziti-router", "dilithium"})
		},
	}
	return nil
}

type kit struct{}

func (_ *kit) Configure(_ *model.Model) error {
	if err := fablib.CopyTree(DilithiumEtc(), filepath.Join(model.KitBuild(), "cfg/dilithium")); err != nil {
		return errors.Wrap(err, "error copying dilithium etc into kit")
	}
	return nil
}
