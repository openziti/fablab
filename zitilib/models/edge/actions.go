package edge

import (
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/zitilib/models/edge/actions"
)

func newActionsFactory() model.Factory {
	return &actionsFactory{}
}

func (f *actionsFactory) Build(m *model.Model) error {
	m.Actions = model.ActionBinders{
		"bootstrap": actions.NewBootstrapAction(),
		"start":     actions.NewStartAction(),
	}
	return nil
}

type actionsFactory struct{}
