package edge

import (
	"errors"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/zitilib/cli"
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

	_, err := cli.Exec(m, "edge", "login", edgeApiBaseUrl, "-c", caChain, "-u", username, "-p", password)
	return err
}

type login struct {
	ctrl *model.Host
}
