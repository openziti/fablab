package edge

import (
	"errors"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/openziti/fablab/kernel/model"
	zitilib_actions "github.com/openziti/fablab/zitilib/actions"
	"io/ioutil"
	"os"
	"path/filepath"
)

func CtrlAddRouters(regionSpec, hostSpec, componentSpec string) model.Action {
	return &ctrlAddRouters{
		regionSpec:    regionSpec,
		hostSpec:      hostSpec,
		componentSpec: componentSpec,
	}
}

func (ar *ctrlAddRouters) Execute(m *model.Model) error {
	hosts := m.SelectHosts(ar.regionSpec, ar.hostSpec)
	for _, h := range hosts {
		components := h.SelectComponents(ar.componentSpec)
		for _, c := range components {
			if c.Data == nil {
				c.Data = model.Data{}
			}

			existingRouter, err := ar.getRouter(m, c)

			if err != nil {
				return err
			}

			if existingRouter.Data() == nil {
				existingRouter, err = ar.createRouter(m, c)
				if err != nil {
					return err
				}
			}

			if existingRouter.Path("enrollmentJwt").Data() == nil {
				c.Data["isEnrolled"] = true
				return nil
			}

			jwt, ok := existingRouter.Path("enrollmentJwt").Data().(string)

			if !ok {
				return fmt.Errorf("could not extract enrollment JWT for edge-router [%s]", c.PublicIdentity)
			}

			localDest := filepath.Join(model.ConfigBuild(), c.PublicIdentity+".jwt")
			remoteDest := "/home/fedora/fablab/" + c.PublicIdentity + ".jwt"

			c.Data["localJwt"] = localDest
			c.Data["remoteJwt"] = remoteDest

			return ioutil.WriteFile(localDest, []byte(jwt), os.ModePerm)
		}
	}
	return nil
}

func (ar *ctrlAddRouters) getRouter(m *model.Model, c *model.Component) (*gabs.Container, error) {
	filter := fmt.Sprintf(`name="%s"`, c.PublicIdentity)
	out, err := zitilib_actions.Edge("edge", "controller", "list", "edge-routers", filter, "-j").ExecuteWithOutput(m)

	if err != nil {
		return nil, err
	}

	data, err := gabs.ParseJSON([]byte(out))
	if err != nil {
		return nil, err
	}

	return data.Path("data").Index(0), nil
}

func (ar *ctrlAddRouters) createRouter(m *model.Model, c *model.Component) (*gabs.Container, error) {
	out, err := zitilib_actions.Edge("edge", "controller", "create", "edge-router", c.PublicIdentity, "-j").ExecuteWithOutput(m)
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
	out, err = zitilib_actions.Edge("edge", "controller", "list", "edge-routers", filter, "-j").ExecuteWithOutput(m)

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

type ctrlAddRouters struct {
	regionSpec    string
	hostSpec      string
	componentSpec string
}
