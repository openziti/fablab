package edge

import (
	"errors"
	"fmt"
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/fablib/actions/host"
	"github.com/openziti/fablab/kernel/model"
)

func InitController(componentSpec string) model.Action {
	return &edgeInit{
		componentSpec: componentSpec,
	}
}

func (init *edgeInit) Execute(m *model.Model) error {
	username := m.MustVariable("credentials", "edge", "username").(string)
	password := m.MustVariable("credentials", "edge", "password").(string)

	if username == "" {
		return errors.New("variable credentials/edge/username must be a string")
	}

	if password == "" {
		return errors.New("variable credentials/edge/password must be a string")
	}

	for _, c := range m.SelectComponents(init.componentSpec) {
		sshConfigFactory := fablib.NewSshConfigFactoryImpl(m, c.GetHost().PublicIp)

		tmpl := "rm -f /home/%v/fablab/ctrl.db && set -o pipefail; /home/%s/fablab/bin/%s --log-formatter pfxlog edge init /home/%s/fablab/cfg/%s -u %s -p %s 2>&1 | tee logs/%s.edge.init.log"
		if err := host.Exec(c.GetHost(), fmt.Sprintf(tmpl, sshConfigFactory.User(), sshConfigFactory.User(), c.BinaryName, sshConfigFactory.User(), c.ConfigName, username, password, c.BinaryName)).Execute(m); err != nil {
			return err
		}
	}

	return nil
}

type edgeInit struct {
	componentSpec string
}
