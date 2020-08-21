package edge

import (
	"errors"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/fablib/actions/host"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/zitilib/cli"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func InitEdgeRouters(regionSpec, hostSpec, componentSpec string) model.Action {
	return &initEdgeRoutersAction{
		regionSpec:    regionSpec,
		hostSpec:      hostSpec,
		componentSpec: componentSpec,
	}
}

func (action *initEdgeRoutersAction) Execute(m *model.Model) error {
	hosts := m.SelectHosts(action.regionSpec, action.hostSpec)
	for _, h := range hosts {
		components := h.SelectComponents(action.componentSpec)
		for _, c := range components {
			if _, err := cli.Exec(m, "edge", "delete", "edge-router", c.PublicIdentity); err != nil {
				return err
			}

			routerCreateResult, err := action.createRouter(m, c)
			if err != nil {
				return err
			}

			jwt, ok := routerCreateResult.Path("enrollmentJwt").Data().(string)

			if !ok {
				return fmt.Errorf("could not extract enrollment JWT for edge-router [%s]", c.PublicIdentity)
			}

			jwtFileName := filepath.Join(model.ConfigBuild(), c.PublicIdentity+".jwt")

			if err := ioutil.WriteFile(jwtFileName, []byte(jwt), os.ModePerm); err != nil {
				return err
			}

			if err := action.enrollRouter(m, h, c, jwtFileName); err != nil {
				return err
			}
		}
	}
	return nil
}

func (action *initEdgeRoutersAction) getRouter(m *model.Model, c *model.Component) (*gabs.Container, error) {
	filter := fmt.Sprintf(`name="%s"`, c.PublicIdentity)
	out, err := cli.Exec(m, "edge", "list", "edge-routers", filter, "-j")

	if err != nil {
		return nil, err
	}

	data, err := gabs.ParseJSON([]byte(out))
	if err != nil {
		return nil, err
	}

	return data.Path("data").Index(0), nil
}

func (action *initEdgeRoutersAction) createRouter(m *model.Model, c *model.Component) (*gabs.Container, error) {
	out, err := cli.Exec(m, "edge", "create", "edge-router", c.PublicIdentity, "-j")
	if err != nil {
		return nil, err
	}

	data, err := gabs.ParseJSON([]byte(out))

	if err != nil {
		return nil, err
	}

	id := data.Path("data.id").Data().(string)

	if id == "" {
		return nil, errors.New("could not obtain edge-router id")
	}

	filter := fmt.Sprintf(`id="%s"`, id)
	out, err = cli.Exec(m, "edge", "list", "edge-routers", filter, "-j")

	if err != nil {
		return nil, err
	}

	data, err = gabs.ParseJSON([]byte(out))

	if err != nil {
		return nil, err
	}

	router := data.Path("data").Index(0)

	if router.Data() == nil {
		return nil, fmt.Errorf("expected edge router with id [%s] to exist", id)
	}

	return router, nil

}

func (action *initEdgeRoutersAction) enrollRouter(m *model.Model, h *model.Host, c *model.Component, jwtFileName string) error {
	ssh := fablib.NewSshConfigFactoryImpl(m, h.PublicIp)

	remoteJwt := "/home/fedora/fablab/" + c.PublicIdentity + ".jwt"
	if err := fablib.SendFile(ssh, jwtFileName, remoteJwt); err != nil {
		return err
	}
	sshConfigFactory := fablib.NewSshConfigFactoryImpl(m, h.PublicIp)
	if output, err := fablib.RemoteExec(sshConfigFactory, "mkdir -p /home/fedora/logs"); err == nil {
		if output != "" {
			logrus.Infof("output [%s]", strings.Trim(output, " \t\r\n"))
		}
	} else {
		return err
	}

	tmpl := "set -o pipefail; /home/fedora/fablab/bin/%s enroll /home/fedora/fablab/cfg/%s -j %s 2>&1 | tee /home/fedora/logs/%s.router.enroll.log "
	return host.Exec(h, fmt.Sprintf(tmpl, c.BinaryName, c.ConfigName, remoteJwt, c.ConfigName)).Execute(m)
}

type initEdgeRoutersAction struct {
	regionSpec    string
	hostSpec      string
	componentSpec string
}
