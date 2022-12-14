/*
	Copyright 2019 NetFoundry Inc.

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

type EndpointSelectorF func(m *model.Model) string

func Iperf(scenarioName string, endpoint EndpointSelectorF, serverHosts, clientHosts string, seconds int, overlay bool) model.OperatingStage {
	return &iperf{
		scenarioName:     scenarioName,
		endpointSelector: endpoint,
		serverHosts:      serverHosts,
		clientHosts:      clientHosts,
		seconds:          seconds,
		overlay:          overlay,
	}
}

func (i *iperf) Operate(run model.Run) error {
	m := run.GetModel()
	serverHosts := m.SelectHosts(i.serverHosts)
	clientHosts := m.SelectHosts(i.clientHosts)

	for _, host := range clientHosts {
		serverHost := serverHosts[0]
		sshServerFactory := lib.NewSshConfigFactory(serverHost)
		go i.runServer(sshServerFactory)
		time.Sleep(10 * time.Second)
		sshClientFactory := lib.NewSshConfigFactory(host)

		if i.overlay == false {
			iperfCmdU := fmt.Sprintf("iperf3 -c %s -p 7001 -t %d -P 1 -b 640M --json", serverHost.PublicIp, i.seconds)
			output, err := lib.RemoteExec(sshClientFactory, iperfCmdU)
			if err == nil {
				logrus.Debugf("output = [%s]", output)
				if summary, err := lib.SummarizeIperf([]byte(output)); err == nil {
					if host.Data == nil {
						host.Data = make(map[string]interface{})
					}
					metricsKey := fmt.Sprintf("iperf_%s_metrics", i.scenarioName)
					host.Data[metricsKey] = summary
				} else {
					return fmt.Errorf("error summarizing client iperf data [%w]", err)
				}
			} else {
				return fmt.Errorf("iperf3 client failure [%s] (%w)", output, err)
			}
		} else {
			iperfCmdO := fmt.Sprintf("iperf3 -c %s -p 7001 -t %d -P 1 -b 640M --json", "iperf.service", i.seconds)
			output, err := lib.RemoteExec(sshClientFactory, iperfCmdO)
			if err == nil {
				logrus.Debugf("output = [%s]", output)
				if summary, err := lib.SummarizeIperf([]byte(output)); err == nil {
					if host.Data == nil {
						host.Data = make(map[string]interface{})
					}
					metricsKey := fmt.Sprintf("iperf_%s_metrics", i.scenarioName)
					host.Data[metricsKey] = summary
				} else {
					return fmt.Errorf("error summarizing client iperf data [%w]", err)
				}
			} else {
				return fmt.Errorf("iperf3 client failure [%s] (%w)", output, err)
			}
		}
		//output, err := lib.RemoteExec(sshClientFactory, iperfCmdU)
		//output, err := lib.RemoteExec(sshClientFactory, iperfCmdO)
		//if err == nil {
		//	logrus.Debugf("output = [%s]", output)
		//	if summary, err := lib.SummarizeIperf([]byte(output)); err == nil {
		//		if host.Data == nil {
		//			host.Data = make(map[string]interface{})
		//		}
		//		metricsKey := fmt.Sprintf("iperf_%s_metrics", i.scenarioName)
		//		host.Data[metricsKey] = summary
		//	} else {
		//		return fmt.Errorf("error summarizing client iperf data [%w]", err)
		//	}
		//} else {
		//	return fmt.Errorf("iperf3 client failure [%s] (%w)", output, err)
		//}
	}
	//fmt.Printf("clientHosts type and value: %T, ", clientHosts)
	//if len(serverHosts) == 1 && len(clientHosts) == 1 {
	//	serverHost := serverHosts[0]
	//	clientHost := clientHosts[0]
	//
	//	sshClientFactory := lib.NewSshConfigFactory(clientHost)
	//	sshServerFactory := lib.NewSshConfigFactory(serverHost)
	//
	//	go i.runServer(sshServerFactory)
	//
	//	time.Sleep(10 * time.Second)
	//
	//	if err := lib.RemoteKillFilter(sshClientFactory, "iperf3", "sudo"); err != nil {
	//		return fmt.Errorf("error killing iperf3 clients (%w)", err)
	//	}
	//
	//	endpoint := i.endpointSelector(run.GetModel())
	//
	//	iperfCmd := fmt.Sprintf("iperf3 -c %s -p 7001 -t %d -P 128 -b 37M --json", endpoint, i.seconds)
	//	output, err := lib.RemoteExec(sshClientFactory, iperfCmd)
	//	if err == nil {
	//		logrus.Debugf("output = [%s]", output)
	//		if summary, err := lib.SummarizeIperf([]byte(output)); err == nil {
	//			if clientHost.Data == nil {
	//				clientHost.Data = make(map[string]interface{})
	//			}
	//			metricsKey := fmt.Sprintf("iperf_%s_metrics", i.scenarioName)
	//			clientHost.Data[metricsKey] = summary
	//		} else {
	//			return fmt.Errorf("error summarizing client iperf data [%w]", err)
	//		}
	//	} else {
	//		return fmt.Errorf("iperf3 client failure [%s] (%w)", output, err)
	//	}
	//
	//} else {
	//	return fmt.Errorf("found [%d] server hosts, and [%d] client hosts, skipping", len(serverHosts), len(clientHosts))
	//}
	return nil
}

func (i *iperf) runServer(factory lib.SshConfigFactory) {
	if err := lib.RemoteKill(factory, "iperf3"); err != nil {
		logrus.Errorf("error killing iperf3 servers (%v)", err)
		return
	}

	//output, err := lib.RemoteExec(factory, "iperf3 -s -p 7001")
	output, err := lib.RemoteExec(factory, "iperf3 -s -p 7001 --one-off")
	if err == nil {
		logrus.Infof("iperf3 server completed")
	} else {
		logrus.Errorf("iperf3 server failure [%s] (%v)", output, err)
	}
}

type iperf struct {
	scenarioName     string
	endpointSelector EndpointSelectorF
	serverHosts      string
	clientHosts      string
	seconds          int
	overlay          bool
}
