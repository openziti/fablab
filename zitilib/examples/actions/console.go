package zitilib_examples_actions

import (
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/zitilib/examples/console"
)

func NewConsoleAction() model.ActionBinder {
	action := &consoleAction{}
	return action.bind
}

func (_ *consoleAction) bind(m *model.Model) model.Action {
	return console.Console()
}

type consoleAction struct{}