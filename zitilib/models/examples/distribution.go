/*
	Copyright 2019 NetFoundry, Inc.

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
	distribution "github.com/netfoundry/fablab/kernel/fablib/runlevel/3_distribution"
	"github.com/netfoundry/fablab/kernel/fablib/runlevel/3_distribution/rsync"
	"github.com/netfoundry/fablab/kernel/model"
)

func newDistributionFactory() model.Factory {
	return &distributionFactory{}
}

func (_ *distributionFactory) Build(m *model.Model) error {
	m.Distribution = model.DistributionBinders{
		func(m *model.Model) model.DistributionStage { return distribution.Locations("*", "@ctrl", "logs") },
		func(m *model.Model) model.DistributionStage { return distribution.Locations("*", "@router", "logs") },
		func(m *model.Model) model.DistributionStage { return distribution.Locations("*", "@loop-listener", "logs")},
		func(m *model.Model) model.DistributionStage { return distribution.Locations("*", "@loop-dialer", "logs")},

		func(m *model.Model) model.DistributionStage { return rsync.Rsync() },
	}
	return nil
}

type distributionFactory struct{}
