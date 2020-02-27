package mattermozt

import "github.com/netfoundry/fablab/kernel/model"

func newActionsFactory() model.Factory {
	return &actionsFactory{}
}

func (f *actionsFactory) Build(m *model.Model) error {
	m.Actions = model.ActionBinders{
		"bootstrap": newBootstrapAction(),
		"start":     newStartAction(),
	}
	return nil
}

type actionsFactory struct{}

