package operation

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
)

func Tcpdump(scenarioName, region, host string, snaplen int) model.OperatingStage {
	return &tcpdump{
		scenarioName: scenarioName,
		region:       region,
		host:         host,
		snaplen:      snaplen,
	}
}

func (t *tcpdump) Operate(m *model.Model) error {
	hosts := m.GetHosts(t.region, t.host)
	if len(hosts) == 1 {
		ssh := fablib.NewSshConfigFactoryImpl(m, hosts[0].PublicIp)

		if err := fablib.RemoteKill(ssh, "tcpdump"); err != nil {
			return fmt.Errorf("error killing tcpdump instances")
		}

		go t.runTcpdump(ssh)

		return nil

	} else {
		return fmt.Errorf("found [%d] hosts", len(hosts))
	}
}

func (t *tcpdump) runTcpdump(ssh fablib.SshConfigFactory) {
	pcapPath, err := ioutil.TempFile("", fmt.Sprintf("%s_*.pcap", t.scenarioName))
	if err != nil {
		logrus.Fatalf("error creating pcap filename (%w)", err)
	}

	output, err := fablib.RemoteExec(ssh, fmt.Sprintf("sudo tcpdump -s %d -w %s", t.snaplen, filepath.Base(pcapPath.Name())))
	if err != nil {
		logrus.Infof("output = [%s]", output)
	}
}

type tcpdump struct {
	scenarioName string
	region       string
	host         string
	snaplen      int
}
