/*
	Copyright 2019 Netfoundry, Inc.

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
	"github.com/netfoundry/fablab/kernel/internal"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"time"
)

func Iperf(seconds int) model.OperatingStage {
	return &iperf{seconds: seconds}
}

func (iperf *iperf) Operate(m *model.Model) error {
	serverHosts := m.GetHosts("@iperf_server", "@iperf_server")
	clientHosts := m.GetHosts("@iperf_client", "@iperf_client")
	if len(serverHosts) == 1 && len(clientHosts) == 1 {
		serverHost := serverHosts[0]
		clientHost := clientHosts[0]
		sshUser := m.MustVariable("credentials", "ssh", "username").(string)
		go iperf.runServer(serverHost, sshUser)

		time.Sleep(10 * time.Second)

		if err := internal.RemoteKill(sshUser, clientHost.PublicIp, "iperf3"); err != nil {
			return fmt.Errorf("error killing iperf3 clients (%w)", err)
		}

		initiator := m.GetHosts("@initiator", "@initiator")[0]
		iperfCmd := fmt.Sprintf("iperf3 -c %s -p 7002 -t %d --json", initiator.PublicIp, iperf.seconds)
		output, err := internal.RemoteExec(sshUser, clientHost.PublicIp, iperfCmd)
		if err == nil {
			if summary, err := internal.SummarizeIperf([]byte(output)); err == nil {
				if clientHost.Data == nil {
					clientHost.Data = make(map[string]interface{})
				}
				clientHost.Data["iperf_metrics"] = summary
			} else {
				logrus.Errorf("error summarizing client iperf data [%w]", err)
			}
		} else {
			logrus.Errorf("iperf3 client failure [%s] (%w)", output, err)
		}

	} else {
		logrus.Warnf("found [%d] server hosts, and [%d] client hosts, skipping", len(serverHosts), len(clientHosts))
	}
	return nil
}

func (iperf *iperf) runServer(h *model.Host, sshUser string) {
	if err := internal.RemoteKill(sshUser, h.PublicIp, "iperf3"); err != nil {
		logrus.Errorf("error killing iperf3 clients (%w)", err)
		return
	}

	output, err := internal.RemoteExec(sshUser, h.PublicIp, "iperf3 -s -p 7001 --one-off --json")
	if err == nil {
		logrus.Infof("iperf3 server completed, output [%s]", output)
	} else {
		logrus.Errorf("iperf3 server failure [%s] (%w)", output, err)
	}
}

type iperf struct {
	seconds int
}
