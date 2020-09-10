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

package dilithium

import (
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/fablib/runlevel/2_kitting/devkit"
	"github.com/openziti/fablab/kernel/model"
	zitilib_bootstrap "github.com/openziti/fablab/zitilib"
	"github.com/pkg/errors"
	"path/filepath"
)

func newConfigurationFactory() model.Factory {
	return &configurationFactory{}
}

func (_ *configurationFactory) Build(m *model.Model) error {
	m.Configuration = model.ConfigurationStages{
		Kit(),
		devkit.DevKit(zitilib_bootstrap.ZitiDistBinaries(), []string{"dilithium"}),
	}
	return nil
}

type configurationFactory struct{}

func Kit() model.ConfigurationStage {
	return &kit{}
}

func (self *kit) Configure(_ model.Run) error {
	if err := fablib.CopyTree(DilithiumEtc(), filepath.Join(model.KitBuild(), "etc")); err != nil {
		return errors.Wrap(err, "error copying dilithium etc into kit")
	}
	return nil
}

type kit struct{}
