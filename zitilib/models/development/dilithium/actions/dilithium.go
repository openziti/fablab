/*
	(c) Copyright NetFoundry, Inc.

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

package dilithium_actions

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
)

type stopDilithiumTunnel struct {
	regionSpec string
	hostSpec   string
}

func StopDilithiumTunnel(regionSpec, hostSpec string) model.Action {
	return &stopDilithiumTunnel{regionSpec, hostSpec}
}

func (self *stopDilithiumTunnel) Execute(m *model.Model) error {
	hosts := m.SelectHosts(self.regionSpec, self.hostSpec)
	for _, host := range hosts {
		ssh := fablib.NewSshConfigFactoryImpl(m, host.PublicIp)
		if err := fablib.RemoteKill(ssh, "dilithium tunnel"); err != nil {
			return errors.Wrap(err, "kill dilithium tunnel")
		}
		if err := fablib.RemoteKill(ssh, "iperf3"); err != nil {
			return errors.Wrap(err, "kill iperf3")
		}
	}
	return nil
}

type startDilithiumTunnelServer struct {
	regionSpec string
	hostSpec   string
}

func StartDilithiumTunnelServer(regionSpec, hostSpec string) model.Action {
	return &startDilithiumTunnelServer{regionSpec, hostSpec}
}

func (self *startDilithiumTunnelServer) Execute(m *model.Model) error {
	hosts := m.SelectHosts(self.regionSpec, self.hostSpec)
	if len(hosts) != 1 {
		return errors.Errorf("expected [1] diltihium tunnel server host, got [%d]", len(hosts))
	}

	serverHost := hosts[0]
	instrument := serverHost.Variables.Must("dilithium", "instrument")
	ssh := fablib.NewSshConfigFactoryImpl(m, serverHost.PublicIp)

	cmd := fmt.Sprintf("nohup fablab/bin/dilithium tunnel server 0.0.0.0:6262 127.0.0.1:2222 -i %s > logs/dilithium-server.log 2>&1 &", instrument)
	if _, err := fablib.RemoteExec(ssh, cmd); err != nil {
		return errors.Wrap(err, "dilithium tunnel server error")
	}

	cmd = fmt.Sprintf("nohup iperf3 -s -p 2222 > logs/iperf3-server.log 2>&1 &")
	if _, err := fablib.RemoteExec(ssh, cmd); err != nil {
		return errors.Wrap(err, "iperf3 server error")
	}

	return nil
}

type startDilithiumTunnelClient struct {
	regionSpec       string
	hostSpec         string
	serverRegionSpec string
	serverHostSpec   string
}

func StartDilithiumTunnelClient(regionSpec, hostSpec, serverRegionSpec, serverHostSpec string) model.Action {
	return &startDilithiumTunnelClient{regionSpec, hostSpec, serverRegionSpec, serverHostSpec}
}

func (self *startDilithiumTunnelClient) Execute(m *model.Model) error {
	clientHosts := m.SelectHosts(self.regionSpec, self.hostSpec)
	if len(clientHosts) != 1 {
		return errors.Errorf("expected [1] dilithium tunnel client host, got [%d]", len(clientHosts))
	}

	serverHosts := m.SelectHosts(self.serverRegionSpec, self.serverHostSpec)
	if len(serverHosts) != 1 {
		return errors.Errorf("expected [1] dilithium tunnel server host, got [%d]", len(serverHosts))
	}

	clientHost := clientHosts[0]
	instrument := clientHost.Variables.Must("dilithium", "instrument")
	ssh := fablib.NewSshConfigFactoryImpl(m, clientHost.PublicIp)
	cmd := fmt.Sprintf("nohup fablab/bin/dilithium tunnel client %s:6262 127.0.0.1:1122 -i %s > logs/dilithium-client.log 2>&1 &", serverHosts[0].PublicIp, instrument)
	if _, err := fablib.RemoteExec(ssh, cmd); err != nil {
		return errors.Wrap(err, "dilithium tunnel client error")
	}

	return nil
}
