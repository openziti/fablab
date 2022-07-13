package distribution

import (
	"os"

	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func DistributeData(hostSpec string, data []byte, dest string) model.DistributionStage {
	return &distData{
		hostSpec: hostSpec,
		data:     data,
		dest:     dest,
	}
}

func (df *distData) Distribute(run model.Run) error {
	return run.GetModel().ForEachHost(df.hostSpec, 25, func(host *model.Host) error {
		ssh := lib.NewSshConfigFactory(host)
		if err := lib.SendData(ssh, df.data, df.dest); err != nil {
			logrus.Errorf("[%s] unable to send data => %s", host.PublicIp, df.dest)
			return err
		}

		if err := lib.Chmod(ssh, df.dest, os.FileMode(0644)); err != nil {
			logrus.Errorf("[%s] unable to send data => %s", host.PublicIp, df.dest)
			return err
		}

		logrus.Infof("[%s] data => %s", host.PublicIp, df.dest)

		return nil
	})
}

type distData struct {
	hostSpec string
	data     []byte
	dest     string
}
