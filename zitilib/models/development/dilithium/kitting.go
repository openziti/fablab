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

func newKittingFactory() model.Factory {
	return &kittingFactory{}
}

func (_ *kittingFactory) Build(m *model.Model) error {
	m.Kitting = model.KittingBinders{
		func(_ *model.Model) model.KittingStage { return Kit() },
		func(_ *model.Model) model.KittingStage {
			return devkit.DevKit(filepath.Join(zitilib_bootstrap.ZitiDistRoot(), "bin"), []string{"dilithium"})
		},
	}
	return nil
}

type kittingFactory struct{}

func Kit() model.KittingStage {
	return &kit{}
}

func (self *kit) Kit(_ *model.Model) error {
	if err := fablib.CopyTree(DilithiumEtc(), filepath.Join(model.KitBuild(), "etc")); err != nil {
		return errors.Wrap(err, "error copying dilithium etc into kit")
	}
	return nil
}

type kit struct{}
