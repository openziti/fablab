/*
	Copyright 2020 NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package operation

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
)

func Tcpdump(scenarioName, region, host string, snaplen int, joiner chan struct{}) model.OperatingStage {
	return &tcpdump{
		scenario: scenarioName,
		region:   region,
		host:     host,
		snaplen:  snaplen,
		joiner:   joiner,
	}
}

func (t *tcpdump) Operate(m *model.Model, _ string) error {
	hosts := m.SelectHosts(fmt.Sprintf("%v > %v", t.region, t.host))
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
	defer func() {
		if t.joiner != nil {
			close(t.joiner)
			logrus.Debug("joiner closed")
		}
	}()

	pcapPath, err := ioutil.TempFile("", fmt.Sprintf("%s_*.pcap", t.scenario))
	if err != nil {
		logrus.Fatalf("error creating pcap filename (%v)", err)
	}

	output, err := fablib.RemoteExec(ssh, fmt.Sprintf("sudo tcpdump -s %d -w %s", t.snaplen, filepath.Base(pcapPath.Name())))
	if err != nil {
		logrus.Infof("output = [%s]", output)
	}
}

type tcpdump struct {
	scenario string
	region   string
	host     string
	snaplen  int
	joiner   chan struct{}
}
