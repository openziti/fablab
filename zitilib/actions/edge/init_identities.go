package edge

import (
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/zitilib/cli"
	"path/filepath"
	"strings"
)

func InitIdentities(componentSpec string, parallel bool) model.Action {
	return &initIdentitiesAction{
		componentSpec: componentSpec,
		parallel:      parallel,
	}
}

func (action *initIdentitiesAction) Execute(m *model.Model) error {
	return m.ForEachComponent(action.componentSpec, action.parallel, func(c *model.Component) error {
		if _, err := cli.Exec(m, "edge", "delete", "identity", c.PublicIdentity); err != nil {
			return err
		}

		return action.createAndEnrollIdentity(c)
	})
}

func (action *initIdentitiesAction) createAndEnrollIdentity(c *model.Component) error {
	ssh := fablib.NewSshConfigFactoryImpl(c.GetModel(), c.GetHost().PublicIp)

	jwtFileName := filepath.Join(model.ConfigBuild(), c.PublicIdentity+".jwt")

	_, err := cli.Exec(c.GetModel(), "edge", "create", "identity", "service", c.PublicIdentity,
		"--jwt-output-file", jwtFileName,
		"-a", strings.Join(c.Tags, ","))

	if err != nil {
		return err
	}

	configFileName := filepath.Join(model.ConfigBuild(), c.PublicIdentity+".json")

	_, err = cli.Exec(c.GetModel(), "edge", "enroll", "--jwt", jwtFileName, "--out", configFileName)

	if err != nil {
		return err
	}

	remoteConfigFile := "/home/fedora/fablab/cfg/" + c.PublicIdentity + ".json"
	return fablib.SendFile(ssh, configFileName, remoteConfigFile)
}

type initIdentitiesAction struct {
	componentSpec string
	parallel      bool
}
