/*
	Copyright NetFoundry Inc.

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
	"github.com/pkg/errors"
)

type iperfClient struct {
	hostSpec string
	address  string
	port     int
}

func IperfClient(hostSpec, address string, port int) model.Stage {
	return &iperfClient{hostSpec, address, port}
}

func (self *iperfClient) Execute(run model.Run) error {
	m := run.GetModel()
	hosts := m.SelectHosts(self.hostSpec)
	if len(hosts) != 1 {
		return errors.Errorf("expected [1] iperf client host, found [%d]", len(hosts))
	}

	ssh := lib.NewSshConfigFactory(hosts[0])

	cmd := fmt.Sprintf("iperf3 -c %s -p %d", self.address, self.port)
	if err := lib.RemoteConsole(ssh, cmd); err != nil {
		return errors.Wrap(err, "iperf3 client exec")
	}

	return nil
}
