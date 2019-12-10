package operation

import (
	"github.com/netfoundry/fablab/kernel"
	"github.com/sirupsen/logrus"
)

func IperfServer() kernel.OperatingStage {
	return &iperfServer{}
}

func (iperfServer *iperfServer) Operate(m *kernel.Model) error {
	iperfHosts := m.GetHosts("@iperf-server", "@iperf-server")
	if len(iperfHosts) == 1 {
		go iperfServer.run(iperfHosts[0], m)

	} else {
		logrus.Warnf("found [%d] iperf hosts, skipping server", len(iperfHosts))
	}
	return nil
}

func (iperfServer *iperfServer) run(h *kernel.Host, m *kernel.Model) {
	logrus.Infof("running iperf server")
}

type iperfServer struct {
}