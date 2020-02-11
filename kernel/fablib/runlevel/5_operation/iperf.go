/*
	Copyright 2019 NetFoundry, Inc.

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
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"time"
)

func Iperf(scenarioName, endpoint, serverRegion, serverHost, clientRegion, clientHost string, seconds int) model.OperatingStage {
	return &iperf{
		scenarioName: scenarioName,
		endpoint:     endpoint,
		serverRegion: serverRegion,
		serverHost:   serverHost,
		clientRegion: clientRegion,
		clientHost:   clientHost,
		seconds:      seconds,
	}
}

func (i *iperf) Operate(m *model.Model, _ string) error {
	serverHosts := m.GetHosts(i.serverRegion, i.serverHost)
	clientHosts := m.GetHosts(i.clientRegion, i.clientHost)
	if len(serverHosts) == 1 && len(clientHosts) == 1 {
		serverHost := serverHosts[0]
		clientHost := clientHosts[0]

		sshClientFactory := fablib.NewSshConfigFactoryImpl(m, clientHost.PublicIp)
		sshServerFactory := fablib.NewSshConfigFactoryImpl(m, serverHost.PublicIp)

		go i.runServer(sshServerFactory)

		time.Sleep(10 * time.Second)

		if err := fablib.RemoteKill(sshClientFactory, "iperf3"); err != nil {
			return fmt.Errorf("error killing iperf3 clients (%w)", err)
		}

		iperfCmd := fmt.Sprintf("iperf3 -c %s -p 7001 -t %d --json", i.endpoint, i.seconds)
		output, err := fablib.RemoteExec(sshClientFactory, iperfCmd)
		if err == nil {
			logrus.Infof("output = [%s]", output)
			if summary, err := fablib.SummarizeIperf([]byte(output)); err == nil {
				if clientHost.Data == nil {
					clientHost.Data = make(map[string]interface{})
				}
				metricsKey := fmt.Sprintf("iperf_%s_metrics", i.scenarioName)
				clientHost.Data[metricsKey] = summary
			} else {
				return fmt.Errorf("error summarizing client iperf data [%w]", err)
			}
		} else {
			return fmt.Errorf("iperf3 client failure [%s] (%w)", output, err)
		}

	} else {
		return fmt.Errorf("found [%d] server hosts, and [%d] client hosts, skipping", len(serverHosts), len(clientHosts))
	}
	return nil
}

func (i *iperf) runServer(factory fablib.SshConfigFactory) {
	if err := fablib.RemoteKill(factory, "iperf3"); err != nil {
		logrus.Errorf("error killing iperf3 servers (%w)", err)
		return
	}

	output, err := fablib.RemoteExec(factory, "iperf3 -s -p 7001 --one-off")
	if err == nil {
		logrus.Infof("iperf3 server completed, output [%s]", output)
	} else {
		logrus.Errorf("iperf3 server failure [%s] (%w)", output, err)
	}
}

type iperf struct {
	scenarioName string
	endpoint     string
	serverRegion string
	serverHost   string
	clientRegion string
	clientHost   string
	seconds      int
}
