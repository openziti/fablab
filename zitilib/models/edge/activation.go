package edge

import (
	"github.com/openziti/fablab/kernel/model"
)

func newActivationFactory() model.Factory {
	return &activationFactory{}
}

func (f *activationFactory) Build(m *model.Model) error {
	m.AddActivationActions("bootstrap", "start")
	return nil
}

type activationFactory struct{}
