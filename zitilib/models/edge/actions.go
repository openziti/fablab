package edge

import (
	actions2 "github.com/openziti/fablab/kernel/fablib/actions"
	"github.com/openziti/fablab/kernel/fablib/actions/component"
	"github.com/openziti/fablab/kernel/fablib/actions/host"
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
		"stop": func(m *model.Model) model.Action {
			return component.StopInParallel("*", 15)
		},
		"syncModelEdgeState": actions.NewSyncModelEdgeStateAction(),
		"clean": func(m *model.Model) model.Action {
			return actions2.Workflow(
				component.StopInParallel("*", 15),
				host.GroupExec("*", 25, "rm -f logs/*"),
			)
		},
	}
	return nil
}

type actionsFactory struct{}
