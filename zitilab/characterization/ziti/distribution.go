package zitilab_characterization_ziti

import (
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/kernel/runlevel/3_distribution/rsync"
)

func newDistributionFactory() model.Factory {
	return &distributionFactory{}
}

func (f *distributionFactory) Build(m *model.Model) error {
	m.Distribution = model.DistributionBinders{
		func(m *model.Model) model.DistributionStage { return rsync.Rsync() },
	}
	return nil
}

type distributionFactory struct{}