package distribution

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Locations(regionSpec, hostSpec string, paths ...string) model.DistributionStage {
	return &locations{
		regionSpec: regionSpec,
		hostSpec:   hostSpec,
		paths:      paths,
	}
}

func (self *locations) Distribute(m *model.Model) error {
	hosts := m.GetHosts(self.regionSpec, self.hostSpec)
	for _, host := range hosts {
		ssh := fablib.NewSshConfigFactoryImpl(m, host.PublicIp)
		for _, path := range self.paths {
			mkdir := fmt.Sprintf("mkdir -p %s", path)
			if _, err := fablib.RemoteExec(ssh, mkdir); err == nil {
				logrus.Infof("%s => %s", host.PublicIp, path)
			} else {
				return fmt.Errorf("error creating path [%s] on host [%s] (%w)", path, host.PublicIp, err)
			}
		}
	}
	return nil
}

type locations struct {
	regionSpec string
	hostSpec   string
	paths      []string
}
