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
	aws_ssh_keys0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/aws_ssh_key"
	semaphore0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/semaphore"
	terraform0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/terraform"
	"github.com/openziti/fablab/kernel/fablib/runlevel/1_configuration/config"
	"github.com/openziti/fablab/kernel/fablib/runlevel/2_kitting/devkit"
	distribution "github.com/openziti/fablab/kernel/fablib/runlevel/3_distribution"
	"github.com/openziti/fablab/kernel/fablib/runlevel/3_distribution/rsync"
	operation "github.com/openziti/fablab/kernel/fablib/runlevel/5_operation"
	aws_ssh_keys6 "github.com/openziti/fablab/kernel/fablib/runlevel/6_disposal/aws_ssh_key"
	terraform6 "github.com/openziti/fablab/kernel/fablib/runlevel/6_disposal/terraform"
	"github.com/openziti/fablab/kernel/model"
	zitilib_bootstrap "github.com/openziti/fablab/zitilib"
	"github.com/openziti/fablab/zitilib/models"
	zitilib_runlevel_1_configuration "github.com/openziti/fablab/zitilib/runlevel/1_configuration"
	zitilib_runlevel_5_operation "github.com/openziti/fablab/zitilib/runlevel/5_operation"
	"github.com/pkg/errors"
	"path/filepath"
	"time"
)

type stagesFactory struct{}

func newStagesFactory() model.Factory {
	return &stagesFactory{}
}

func (self *stagesFactory) Build(m *model.Model) error {
	m.Infrastructure = model.InfrastructureStages{
		aws_ssh_keys0.Express(),
		terraform0.Express(),
		semaphore0.Restart(90 * time.Second),
	}

	m.Configuration = model.ConfigurationStages{
		zitilib_runlevel_1_configuration.IfNoPki(zitilib_runlevel_1_configuration.Fabric(), zitilib_runlevel_1_configuration.DotZiti()),
		config.Component(),
		&kit{},
		devkit.DevKit(zitilib_bootstrap.ZitiDistBinaries(), []string{"ziti-controller", "ziti-router", "dilithium"}),
	}

	m.Distribution = model.DistributionStages{
		distribution.Locations("*", "logs"),
		rsync.Rsync(),
	}

	m.AddActivationActions("bootstrap", "start")

	directEndpoint := m.MustSelectHost(models.RemoteId).PublicIp
	remoteProxy := m.MustSelectHost(models.LocalId).PrivateIp

	c := make(chan struct{})
	m.Operation = model.OperatingStages{
		zitilib_runlevel_5_operation.Metrics(c),

		operation.Banner("transwarp"),
		operation.Iperf("ziti", remoteProxy, models.RemoteId, models.LocalId, 30),
		operation.Persist(),

		operation.Banner("internet"),
		operation.Iperf("internet", directEndpoint, models.RemoteId, models.LocalId, 30),
		operation.Persist(),

		operation.Closer(c),
		operation.Persist(),
	}

	m.Disposal = model.DisposalStages{
		terraform6.Dispose(),
		aws_ssh_keys6.Dispose(),
	}

	return nil
}

type kit struct{}

func (_ *kit) Configure(_ model.Run) error {
	if err := fablib.CopyTree(DilithiumEtc(), filepath.Join(model.KitBuild(), "cfg/dilithium")); err != nil {
		return errors.Wrap(err, "error copying dilithium etc into kit")
	}
	return nil
}
