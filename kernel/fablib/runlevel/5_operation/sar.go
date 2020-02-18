package operation

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Sar(host *model.Host, intervalSeconds, snapshots int) model.OperatingStage {
	return &sar{
		host:            host,
		intervalSeconds: intervalSeconds,
		snapshots:       snapshots,
	}
}

func (s *sar) Operate(m *model.Model, _ string) error {
	logrus.Infof("ip = %s", s.host.PublicIp)
	ssh := fablib.NewSshConfigFactoryImpl(m, s.host.PublicIp)
	go s.runSar(ssh)
	return nil
}

func (s *sar) runSar(ssh fablib.SshConfigFactory) {
	sar := fmt.Sprintf("sudo sar -u -r -q %d %d", s.intervalSeconds, s.snapshots)
	output, err := fablib.RemoteExec(ssh, sar)
	if err == nil {
		logrus.Infof("sar completed [%s]", output)
	} else {
		logrus.Errorf("sar failed [%s] (%w)", output, err)
	}
}

type sar struct {
	host            *model.Host
	intervalSeconds int
	snapshots       int
}
