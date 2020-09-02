package edge

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/fablib/actions/host"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/zitilib/cli"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

func InitEdgeRouters(componentSpec string) model.Action {
	return &initEdgeRoutersAction{
		componentSpec: componentSpec,
	}
}

func (action *initEdgeRoutersAction) Execute(m *model.Model) error {
	for _, c := range m.SelectComponents(action.componentSpec) {
		if _, err := cli.Exec(m, "edge", "delete", "edge-router", c.PublicIdentity); err != nil {
			return err
		}

		if err := action.createAndEnrollRouter(c); err != nil {
			return err
		}
	}

	return nil
}

func (action *initEdgeRoutersAction) createAndEnrollRouter(c *model.Component) error {
	ssh := fablib.NewSshConfigFactoryImpl(c.GetModel(), c.GetHost().PublicIp)

	jwtFileName := filepath.Join(model.ConfigBuild(), c.PublicIdentity+".jwt")

	_, err := cli.Exec(c.GetModel(), "edge", "create", "edge-router", c.PublicIdentity, "-j",
		"--jwt-output-file", jwtFileName,
		"-a", strings.Join(c.Tags, ","))

	if err != nil {
		return err
	}

	remoteJwt := "/home/fedora/fablab/" + c.PublicIdentity + ".jwt"
	if err := fablib.SendFile(ssh, jwtFileName, remoteJwt); err != nil {
		return err
	}
	sshConfigFactory := fablib.NewSshConfigFactoryImpl(c.GetModel(), c.GetHost().PublicIp)
	if output, err := fablib.RemoteExec(sshConfigFactory, "mkdir -p /home/fedora/logs"); err == nil {
		if output != "" {
			logrus.Infof("output [%s]", strings.Trim(output, " \t\r\n"))
		}
	} else {
		return err
	}

	tmpl := "set -o pipefail; /home/fedora/fablab/bin/%s enroll /home/fedora/fablab/cfg/%s -j %s 2>&1 | tee /home/fedora/logs/%s.router.enroll.log "
	return host.Exec(c.GetHost(), fmt.Sprintf(tmpl, c.BinaryName, c.ConfigName, remoteJwt, c.ConfigName)).Execute(c.GetModel())
}

type initEdgeRoutersAction struct {
	componentSpec string
}
