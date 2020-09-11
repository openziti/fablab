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
	aws_ssh_keys0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/aws_ssh_key"
	semaphore0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/semaphore"
	terraform0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/terraform"
	"github.com/openziti/fablab/kernel/fablib/runlevel/2_kitting/devkit"
	distribution "github.com/openziti/fablab/kernel/fablib/runlevel/3_distribution"
	"github.com/openziti/fablab/kernel/fablib/runlevel/3_distribution/rsync"
	aws_ssh_keys6 "github.com/openziti/fablab/kernel/fablib/runlevel/6_disposal/aws_ssh_key"
	terraform6 "github.com/openziti/fablab/kernel/fablib/runlevel/6_disposal/terraform"
	"github.com/openziti/fablab/kernel/model"
	zitilib_bootstrap "github.com/openziti/fablab/zitilib"
	"github.com/pkg/errors"
	"path/filepath"
	"time"
)

func newStagesFactory() model.Factory {
	return &stagesFactory{}
}

func (stagesFactory) Build(m *model.Model) error {
	m.Infrastructure = model.InfrastructureStages{
		aws_ssh_keys0.Express(),
		terraform0.Express(),
		semaphore0.Restart(90 * time.Second),
	}

	m.Configuration = model.ConfigurationStages{
		&kit{},
		devkit.DevKit(zitilib_bootstrap.ZitiDistBinaries(), []string{"dilithium"}),
	}

	m.Distribution = model.DistributionStages{
		distribution.Locations("#host", "logs"),
		rsync.Sequential(),
	}

	m.Disposal = model.DisposalStages{
		terraform6.Dispose(),
		aws_ssh_keys6.Dispose(),
	}
	return nil
}

type stagesFactory struct{}

func (self *kit) Configure(model.Run) error {
	if err := fablib.CopyTree(DilithiumEtc(), filepath.Join(model.KitBuild(), "etc")); err != nil {
		return errors.Wrap(err, "error copying dilithium etc into kit")
	}
	return nil
}

type kit struct{}
