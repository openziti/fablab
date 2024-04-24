package distribution

import (
	"fmt"
	"github.com/openziti/fablab/kernel/libssh"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func DistributeSshKey(hostSpec string) model.Stage {
	return &distSshKey{
		hostSpec: hostSpec,
	}
}

func (self *distSshKey) Execute(run model.Run) error {
	return run.GetModel().ForEachHost(self.hostSpec, 25, func(host *model.Host) error {
		ssh := host.NewSshConfigFactory()
		keyPath := fmt.Sprintf("/home/%v/.ssh/id_rsa", ssh.User())

		if _, err := libssh.RemoteExecAll(ssh, fmt.Sprintf("rm -f %v", keyPath)); err == nil {
			logrus.Infof("%s => %s", host.PublicIp, "removed old PK")
		} else {
			return fmt.Errorf("error removing old PK on host [%s] (%w)", host.PublicIp, err)
		}

		if err := libssh.SendFile(ssh, ssh.KeyPath(), keyPath); err != nil {
			logrus.Errorf("[%s] unable to send %s => %s", host.PublicIp, ssh.KeyPath(), keyPath)
			return fmt.Errorf("[%s] unable to send %s => %s (%w)", host.PublicIp, ssh.KeyPath(), keyPath, err)
		}

		logrus.Infof("[%s] %s => %s", host.PublicIp, ssh.KeyPath(), keyPath)

		if _, err := libssh.RemoteExecAll(ssh, fmt.Sprintf("chmod 0400 %v", keyPath)); err == nil {
			logrus.Infof("%s => %s", host.PublicIp, "set pk permissions")
			return nil
		} else {
			return fmt.Errorf("error setting pk permissions on host [%s] (%w)", host.PublicIp, err)
		}
	})
}

type distSshKey struct {
	hostSpec string
}
