package zitilab_characterization_ziti

import (
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/kernel/runlevel/2_kitting/devkit"
	"github.com/netfoundry/fablab/zitilab/development/bootstrap"
	"path/filepath"
)

func newKittingFactory() model.Factory {
	return &kittingFactory{}
}

func (f *kittingFactory) Build(m *model.Model) error {
	m.Kitting = model.KittingBinders{
		func(m *model.Model) model.KittingStage {
			zitiBinaries := []string{
				"ziti-controller",
				"ziti-fabric",
				"ziti-fabric-test",
				"ziti-router",
			}
			return devkit.DevKit(filepath.Join(zitilab_bootstrap.ZitiRoot(), "bin"), zitiBinaries)
		},
	}
	return nil
}

type kittingFactory struct{}
