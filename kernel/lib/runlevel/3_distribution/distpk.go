package distribution

import (
	"fmt"
	"time"

	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

const (
	defaultDistRetries      = 3
	defaultDistRetryBackoff = 5 * time.Second
)

func retryOnHost(host *model.Host, f func() error) error {
	var lastErr error
	for attempt := 1; attempt <= defaultDistRetries; attempt++ {
		lastErr = f()
		if lastErr == nil {
			return nil
		}
		if attempt < defaultDistRetries {
			logrus.WithError(lastErr).Warnf("distribution to host [%s] failed (attempt %d/%d), retrying in %v",
				host.PublicIp, attempt, defaultDistRetries, defaultDistRetryBackoff)
			time.Sleep(defaultDistRetryBackoff)
		}
	}
	return lastErr
}

func DistributeSshKey(hostSpec string) model.Stage {
	return &distSshKey{
		hostSpec: hostSpec,
	}
}

func (self *distSshKey) Execute(run model.Run) error {
	return run.GetModel().ForEachHost(self.hostSpec, 25, func(host *model.Host) error {
		return retryOnHost(host, func() error {
			keyPath := fmt.Sprintf("/home/%v/.ssh/id_rsa", host.GetSshUser())
			sshKeyPath := host.NewSshConfigFactory().KeyPath()

			if _, err := host.ExecLogged(fmt.Sprintf("rm -f %v", keyPath)); err != nil {
				return fmt.Errorf("error removing old PK on host [%s] (%w)", host.PublicIp, err)
			}
			logrus.Infof("%s => %s", host.PublicIp, "removed old PK")

			if err := host.SendFile(sshKeyPath, keyPath); err != nil {
				return fmt.Errorf("[%s] unable to send %s => %s (%w)", host.PublicIp, sshKeyPath, keyPath, err)
			}
			logrus.Infof("[%s] %s => %s", host.PublicIp, sshKeyPath, keyPath)

			if _, err := host.ExecLogged(fmt.Sprintf("chmod 0400 %v", keyPath)); err != nil {
				return fmt.Errorf("error setting pk permissions on host [%s] (%w)", host.PublicIp, err)
			}
			logrus.Infof("%s => %s", host.PublicIp, "set pk permissions")

			return nil
		})
	})
}

type distSshKey struct {
	hostSpec string
}
