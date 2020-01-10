package zitilab_characterization_ziti

import (
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/kernel/runlevel/4_activation/action"
)

func newActivationFactory() model.Factory {
	return &activationFactory{}
}

func (f *activationFactory) Build(m *model.Model) error {
	m.Activation = model.ActivationBinders{
		func(m *model.Model) model.ActivationStage {
			return action.Activation("bootstrap", "start")
		},
	}
	return nil
}

type activationFactory struct{}
