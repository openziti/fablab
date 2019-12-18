package pki

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"os"
)

func Fabric() model.ConfigurationStage {
	return &fabric{}
}

func (f *fabric) Configure(m *model.Model) error {
	if err := generateCa(); err != nil {
		return fmt.Errorf("error generating ca (%s)", err)
	}
	for regionId, region := range m.Regions {
		for hostId, host := range region.Hosts {
			for componentId, component := range host.Components {
				if component.PublicIdentity != "" {
					logrus.Infof("generating public ip identity [%s/%s] on [%s/%s]", componentId, component.PublicIdentity, regionId, hostId)
					if err := generateCert(component.PublicIdentity, host.PublicIp); err != nil {
						return fmt.Errorf("error generating public identity [%s/%s]", componentId, component.PublicIdentity)
					}
				}
				if component.PrivateIdentity != "" {
					logrus.Infof("generating private ip identity [%s/%s] on [%s/%s]", componentId, component.PrivateIdentity, regionId, hostId)
					if err := generateCert(component.PrivateIdentity, host.PrivateIp); err != nil {
						return fmt.Errorf("error generating private identity [%s/%s]", componentId, component.PrivateIdentity)
					}
				}
			}
		}
	}
	return nil
}

func hasExisitingPki() (bool, error) {
	if _, err := os.Stat(model.PkiBuild()); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return true, err
	}
	return true, nil
}

type fabric struct {
}
