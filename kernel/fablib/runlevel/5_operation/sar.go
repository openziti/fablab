package operation

import (
	"encoding/json"
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Sar(scenario string, host *model.Host, intervalSeconds int, joiner chan struct{}) model.OperatingStage {
	return &sar{
		scenario:        scenario,
		host:            host,
		intervalSeconds: intervalSeconds,
		joiner:          joiner,
	}
}

func SarCloser(host *model.Host) model.OperatingStage {
	return &sarCloser{
		host: host,
	}
}

func (s *sar) Operate(m *model.Model, _ string) error {
	ssh := fablib.NewSshConfigFactoryImpl(m, s.host.PublicIp)
	go s.runSar(ssh)
	return nil
}

func (s *sarCloser) Operate(m *model.Model, _ string) error {
	ssh := fablib.NewSshConfigFactoryImpl(m, s.host.PublicIp)
	if err := fablib.RemoteKill(ssh, "sar"); err != nil {
		return fmt.Errorf("error closing sar (%w)", err)
	}
	return nil
}

func (s *sar) runSar(ssh fablib.SshConfigFactory) {
	defer func() {
		close(s.joiner)
		logrus.Debugf("joiner closed")
	}()

	sar := fmt.Sprintf("sar -u -r -q %d", s.intervalSeconds)
	output, err := fablib.RemoteExec(ssh, sar)
	if err != nil {
		logrus.Warnf("sar exited (%w)", err)
	}

	summary, err := fablib.SummarizeSar([]byte(output))
	if err != nil {
		logrus.Errorf("sar summary failed (%w) [%s]", err, output)
		return
	}
	j, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		logrus.Errorf("error marshaling summary (%w)", err)
		return
	}
	logrus.Debugf("summary = %s", j)

	if s.host.Data == nil {
		s.host.Data = make(model.Data)
	}
	v, found := s.host.Data["host"]
	if !found {
		v = make(model.Data)
		s.host.Data["host"] = v
	}
	v.(model.Data)[s.scenario] = summary

	logrus.Infof("sar data added to host")
}

type sar struct {
	scenario        string
	host            *model.Host
	intervalSeconds int
	joiner          chan struct{}
}

type sarCloser struct {
	host *model.Host
}
