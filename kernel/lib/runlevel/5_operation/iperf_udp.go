/*
	Copyright 2020 NetFoundry Inc.

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
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"time"
)

func IperfUdp(scenarioName, endpoint, serverHosts, clientHosts, bandwidth string, seconds int) model.Stage {
	return &iperfUdp{
		scenarioName: scenarioName,
		endpoint:     endpoint,
		serverHosts:  serverHosts,
		clientHosts:  clientHosts,
		bandwidth:    bandwidth,
		seconds:      seconds,
	}
}

func (i *iperfUdp) Execute(run model.Run) error {
	m := run.GetModel()
	serverHosts := m.SelectHosts(i.serverHosts)
	clientHosts := m.SelectHosts(i.clientHosts)
	if len(serverHosts) == 1 && len(clientHosts) == 1 {
		serverHost := serverHosts[0]
		clientHost := clientHosts[0]

		sshClientFactory := lib.NewSshConfigFactory(clientHost)
		sshServerFactory := lib.NewSshConfigFactory(serverHost)

		go i.runServer(sshServerFactory)

		time.Sleep(10 * time.Second)

		if err := lib.RemoteKill(sshClientFactory, "iperf3"); err != nil {
			return fmt.Errorf("error killing iperf3 clients (%w)", err)
		}

		iperfCmd := fmt.Sprintf("iperf3 -c %s -p 7001 -u -t %d --json", i.endpoint, i.seconds)
		output, err := lib.RemoteExec(sshClientFactory, iperfCmd)
		if err == nil {
			if summary, err := lib.SummarizeIperfUdp([]byte(output)); err == nil {
				if clientHost.Data == nil {
					clientHost.Data = make(map[string]interface{})
				}
				metricsKey := fmt.Sprintf("iperf_udp_%s_metrics", i.scenarioName)
				clientHost.Data[metricsKey] = summary
			} else {
				return fmt.Errorf("error summarizing client i data [%w]", err)
			}
			logrus.Debugf("output = [%s]", output)
		} else {
			return fmt.Errorf("iperf3 client failure [%s] (%w)", output, err)
		}

	} else {
		return fmt.Errorf("found [%d] server hosts, and [%d] client hosts, skipping", len(serverHosts), len(clientHosts))
	}
	return nil
}

func (i *iperfUdp) runServer(factory lib.SshConfigFactory) {
	if err := lib.RemoteKill(factory, "iperf3"); err != nil {
		logrus.Errorf("error killing iperf3 clients (%v)", err)
		return
	}

	output, err := lib.RemoteExec(factory, "iperf3 -s -p 7001 --one-off")
	if err == nil {
		logrus.Infof("iperf3 server completed")
	} else {
		logrus.Errorf("iperf3 server failure [%s] (%v)", output, err)
	}
}

type iperfUdp struct {
	scenarioName string
	endpoint     string
	serverHosts  string
	clientHosts  string
	bandwidth    string
	seconds      int
}
