package edge

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/fablib/actions/host"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/zitilib/cli"
	"path/filepath"
	"strings"
)

func InitEdgeRouters(componentSpec string, parallel bool) model.Action {
	return &initEdgeRoutersAction{
		componentSpec: componentSpec,
		parallel:      parallel,
	}
}

func (action *initEdgeRoutersAction) Execute(m *model.Model) error {
	return m.ForEachComponent(action.componentSpec, action.parallel, func(c *model.Component) error {
		if _, err := cli.Exec(m, "edge", "delete", "edge-router", c.PublicIdentity); err != nil {
			return err
		}

		return action.createAndEnrollRouter(c)
	})
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

	remoteJwt := "/home/fedora/fablab/cfg/" + c.PublicIdentity + ".jwt"
	if err := fablib.SendFile(ssh, jwtFileName, remoteJwt); err != nil {
		return err
	}

	tmpl := "set -o pipefail; /home/fedora/fablab/bin/%s enroll /home/fedora/fablab/cfg/%s -j %s 2>&1 | tee /home/fedora/logs/%s.router.enroll.log "
	return host.Exec(c.GetHost(),
		"mkdir -p /home/fedora/logs",
		fmt.Sprintf(tmpl, c.BinaryName, c.ConfigName, remoteJwt, c.ConfigName)).Execute(c.GetModel())
}

type initEdgeRoutersAction struct {
	componentSpec string
	parallel      bool
}
