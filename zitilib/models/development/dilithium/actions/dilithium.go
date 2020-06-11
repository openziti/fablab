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
	hosts := m.GetHosts(self.regionSpec, self.hostSpec)
	for _, host := range hosts {
		ssh := fablib.NewSshConfigFactoryImpl(m, host.PublicIp)
		if err := fablib.RemoteKill(ssh, "dilithium tunnel"); err != nil {
			return errors.Wrap(err, "kill dilithium tunnel")
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
	hosts := m.GetHosts(self.regionSpec, self.hostSpec)
	if len(hosts) != 1 {
		return errors.Errorf("expected [1] diltihium tunnel server host, got [%d]", len(hosts))
	}

	ssh := fablib.NewSshConfigFactoryImpl(m, hosts[0].PublicIp)
	cmd := fmt.Sprintf("nohup fablab/bin/dilithium tunnel server 0.0.0.0:6262 127.0.0.1:2222 > logs/dilithium-server.log 2>&1 &")
	if _, err := fablib.RemoteExec(ssh, cmd); err != nil {
		return errors.Wrap(err, "dilithium tunnel server error")
	}

	return nil
}

type stasrtDilithiumTunnelClient struct {
	regionSpec       string
	hostSpec         string
	serverRegionSpec string
	serverHostSpec   string
}

func StartDilithiumTunnelClient(regionSpec, hostSpec, serverRegionSpec, serverHostSpec string) model.Action {
	return &stasrtDilithiumTunnelClient{regionSpec, hostSpec, serverRegionSpec, serverHostSpec}
}

func (self *stasrtDilithiumTunnelClient) Execute(m *model.Model) error {
	clientHosts := m.GetHosts(self.regionSpec, self.hostSpec)
	if len(clientHosts) != 1 {
		return errors.Errorf("expected [1] dilithium tunnel client host, got [%d]", len(clientHosts))
	}

	serverHosts := m.GetHosts(self.serverRegionSpec, self.serverHostSpec)
	if len(serverHosts) != 1 {
		return errors.Errorf("expected [1] dilithium tunnel server host, got [%d]", len(serverHosts))
	}

	ssh := fablib.NewSshConfigFactoryImpl(m, clientHosts[0].PublicIp)
	cmd := fmt.Sprintf("nohup fablab/bin/dilithium tunnel client %s:6262 127.0.0.1:1122 > logs/dilithium-client.log 2>&1 &", serverHosts[0].PublicIp)
	if _, err := fablib.RemoteExec(ssh, cmd); err != nil {
		return errors.Wrap(err, "dilithium tunnel client error")
	}

	return nil
}
