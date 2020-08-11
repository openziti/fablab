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
	"github.com/openziti/fablab/kernel/fablib/runlevel/2_kitting/devkit"
	"github.com/openziti/fablab/kernel/model"
	zitilib_bootstrap "github.com/openziti/fablab/zitilib"
	"github.com/pkg/errors"
	"path/filepath"
)

type kittingFactory struct{}

func newKittingFactory() model.Factory {
	return &kittingFactory{}
}

func (_ *kittingFactory) Build(m *model.Model) error {
	m.Kitting = model.KittingBinders{
		func(_ *model.Model) model.KittingStage {
			return &kit{}
		},
		func(_ *model.Model) model.KittingStage {
			return devkit.DevKit(zitilib_bootstrap.ZitiDistBinaries(), []string{"ziti-controller", "ziti-router", "dilithium"})
		},
	}
	return nil
}

type kit struct{}

func (_ *kit) Kit(_ *model.Model) error {
	if err := fablib.CopyTree(DilithiumEtc(), filepath.Join(model.KitBuild(), "cfg/dilithium")); err != nil {
		return errors.Wrap(err, "error copying dilithium etc into kit")
	}
	return nil
}
