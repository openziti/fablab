package operation

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Sar(closer chan struct{}, host *model.Host, intervalSeconds, snapshots int) model.OperatingStage {
	return &sar{
		closer:          closer,
		host:            host,
		intervalSeconds: intervalSeconds,
		snapshots:       snapshots,
		closed:          false,
	}
}

func (s *sar) Operate(m *model.Model, _ string) error {
	logrus.Infof("ip = %s", s.host.PublicIp)
	ssh := fablib.NewSshConfigFactoryImpl(m, s.host.PublicIp)
	go s.waitClose()
	go s.runSar(ssh)
	return nil
}

func (s *sar) waitClose() {
	defer logrus.Infof("closed")
	select {
	case <-s.closer:
		s.closed = true
	}
}

func (s *sar) runSar(ssh fablib.SshConfigFactory) {
	defer logrus.Infof("stopping")
	for !s.closed {
		sar := fmt.Sprintf("sudo sar -u -r -q %d %d", s.intervalSeconds, s.snapshots)
		output, err := fablib.RemoteExec(ssh, sar)
		if err != nil {
			logrus.Errorf("sar failed [%s] (%w)", output, err)
		}

		summary, err := fablib.SummarizeSar([]byte(output))
		if err != nil {
			logrus.Errorf("sar summary failed (%w)", err)
		}

		if s.host.Data == nil {
			s.host.Data = make(model.Data)
		}
		a, found := s.host.Data["host"]
		if !found {
			a = make([]*model.HostSummary, 0)
			s.host.Data["host"] = a
		}
		a = append(a.([]*model.HostSummary), summary)
		s.host.Data["host"] = a
	}
}

type sar struct {
	closer          chan struct{}
	host            *model.Host
	intervalSeconds int
	snapshots       int
	closed          bool
}
