package operation

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func TcpdumpCloser(region, host string) model.OperatingStage {
	return &tcpdumpCloser{
		region: region,
		host:   host,
	}
}

func (t *tcpdumpCloser) Operate(m *model.Model) error {
	hosts := m.GetHosts(t.region, t.host)
	var ssh fablib.SshConfigFactory
	if len(hosts) == 1 {
		ssh = fablib.NewSshConfigFactoryImpl(m, hosts[0].PublicIp)
	} else {
		return fmt.Errorf("found [%d] hosts", len(hosts))
	}

	if err := fablib.RemoteKill(ssh, "tcpdump"); err != nil {
		return fmt.Errorf("error closing tcpdump (%w)", err)
	}
	logrus.Infof("tcpdump closed")
	return nil
}

type tcpdumpCloser struct {
	region string
	host   string
}
