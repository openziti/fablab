package operation

import (
	"github.com/netfoundry/fablab/kernel"
	"github.com/sirupsen/logrus"
)

func IperfClient() kernel.OperatingStage {
	return &iperfClient{}
}

func (iperfClient* iperfClient) Operate(m *kernel.Model) error {
	iperfHosts := m.GetHosts("@iperf-client", "@iperf-client")
	if len(iperfHosts) == 1 {
		go iperfClient.run(iperfHosts[0], m)

	} else {
		logrus.Warnf("found [%d] iperf hosts, skipping client", len(iperfHosts))
	}
	return nil
}

func (iperfClient *iperfClient) run(h *kernel.Host, m *kernel.Model) {
	logrus.Infof("running iperf client")
}

type iperfClient struct{
}