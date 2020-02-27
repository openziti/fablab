package edge

import (
	"errors"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/zitilib/actions/cli"
	"path/filepath"
)

func Login(ctrl *model.Host) model.Action {
	return &login{
		ctrl: ctrl,
	}
}

func (l *login) Execute(m *model.Model) error {

	username := m.MustVariable("credentials", "edge", "username").(string)
	password := m.MustVariable("credentials", "edge", "password").(string)
	edgeApiBaseUrl := l.ctrl.PublicIp + ":1280"

	caChain := filepath.Join(model.PkiBuild(), "intermediate", "certs", "intermediate.cert")

	if username == "" {
		return errors.New("variable credentials/edge/username must be a string")
	}

	if password == "" {
		return errors.New("variable credentials/edge/password must be a string")
	}

	return cli.Edge("edge", "controller", "login", edgeApiBaseUrl, "-c", caChain, "-u", username, "-p", password).Execute(m)
}

type login struct {
	ctrl *model.Host
}
