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

package zitilib_characterization

import (
	"fmt"
	operation "github.com/openziti/fablab/kernel/fablib/runlevel/5_operation"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/zitilib/models"
	"github.com/openziti/fablab/zitilib/runlevel/5_operation"
	"time"
)

func newOperationFactory() model.Factory {
	return &operationFactory{}
}

func (f *operationFactory) Build(m *model.Model) error {
	values := m.SelectHosts("#local > #service")
	var directEndpoint string
	if len(values) == 1 {
		directEndpoint = values[0].PublicIp
	} else {
		return fmt.Errorf("need single host for #local > #service, found [%d]", len(values))
	}

	values = m.SelectHosts("#short > #short")
	var shortProxy string
	if len(values) == 1 {
		shortProxy = values[0].PrivateIp
	} else {
		return fmt.Errorf("need single host for #short > #short, found [%d]", len(values))
	}

	values = m.SelectHosts("#medium > #medium")
	var mediumProxy string
	if len(values) == 1 {
		mediumProxy = values[0].PrivateIp
	} else {
		return fmt.Errorf("need a single host for #medium > #medium, found [%d]", len(values))
	}

	values = m.SelectHosts("#long > #long")
	var longProxy string
	if len(values) == 1 {
		longProxy = values[0].PrivateIp
	} else {
		return fmt.Errorf("need a single host for #long > #long, found [%d]", len(values))
	}

	minutes := m.Variables.Must("characterization", "sample_minutes")
	seconds := int((time.Duration(minutes.(int)) * time.Minute).Seconds())

	tcpdump := m.Variables.Must("characterization", "tcpdump", "enabled").(bool)
	snaplen := m.Variables.Must("characterization", "tcpdump", "snaplen").(int)

	c := make(chan struct{})
	m.AddOperatingStages(zitilib_runlevel_5_operation.Mesh(c), zitilib_runlevel_5_operation.Metrics(c))

	f.addStagesForRegion("#short", shortProxy, directEndpoint, tcpdump, snaplen, seconds, m)
	f.addStagesForRegion("#medium", mediumProxy, directEndpoint, tcpdump, snaplen, seconds, m)
	f.addStagesForRegion("#long", longProxy, directEndpoint, tcpdump, snaplen, seconds, m)

	m.AddOperatingStages(operation.Closer(c), operation.Persist())

	return nil
}

func (f *operationFactory) addStagesForRegion(region, initiatingRouter, directEndpoint string, tcpdump bool, snaplen, seconds int, m *model.Model) {
	serverHosts := model.Selector(models.LocalId, models.ServiceTag)
	clientHosts := model.Selector(region, models.ClientTag)

	/*
	 * Ziti Bandwidth Testing
	 */
	scenario0 := fmt.Sprintf("%s_ziti", region)
	m.AddOperatingStage(operation.Banner(scenario0))
	joiners0 := f.sarStages(scenario0, m, 1)

	if tcpdump {
		joiner := make(chan struct{})
		m.AddOperatingStage(operation.Tcpdump("ziti", clientHosts, snaplen, joiner))
		joiners0 = append(joiners0, joiner)
	}
	m.AddOperatingStage(operation.Iperf("ziti", initiatingRouter, serverHosts, clientHosts, seconds))

	if tcpdump {
		m.AddOperatingStage(operation.TcpdumpCloser(clientHosts))
	}
	m.AddOperatingStage(operation.Persist())
	f.sarCloserStages(m)
	m.AddOperatingStage(operation.Joiner(joiners0))
	/* */

	/*
	 * Internet Bandwidth Testing
	 */
	scenario1 := fmt.Sprintf("%s_internet", region)
	m.AddOperatingStage(operation.Banner(scenario1))
	joiners1 := f.sarStages(scenario1, m, 1)

	if tcpdump {
		joiner := make(chan struct{})
		m.AddOperatingStage(operation.Tcpdump("internet", clientHosts, snaplen, joiner))
		joiners1 = append(joiners1, joiner)
	}
	m.AddOperatingStage(operation.Iperf("internet", directEndpoint, serverHosts, clientHosts, seconds))

	if tcpdump {
		m.AddOperatingStage(operation.TcpdumpCloser(clientHosts))
	}
	m.AddOperatingStage(operation.Persist())
	f.sarCloserStages(m)
	m.AddOperatingStage(operation.Joiner(joiners1))
	/* */

	/*
	 * Retrieve tcpdump Captures
	 */
	if tcpdump {
		m.AddOperatingStage(operation.Retrieve(clientHosts, ".", ".pcap"))
	}
	/* */

	/*
	 * Ziti UDP Testing
	 */
	scenario2 := fmt.Sprintf("%s_ziti_udp", region)
	m.AddOperatingStage(operation.Banner(scenario2))
	joiners2 := f.sarStages(scenario2, m, 1)
	m.AddOperatingStage(operation.IperfUdp("ziti_1m", initiatingRouter, serverHosts, clientHosts, "1M", seconds))
	f.sarCloserStages(m)
	m.AddOperatingStage(operation.Joiner(joiners2))
	/* */

	/*
	 * Internet UDP Testing
	 */
	scenario3 := fmt.Sprintf("%s_internet_udp", region)
	m.AddOperatingStage(operation.Banner(scenario3))
	joiners3 := f.sarStages(scenario3, m, 1)
	m.AddOperatingStage(operation.IperfUdp("internet_1m", directEndpoint, serverHosts, clientHosts, "1M", seconds))
	f.sarCloserStages(m)
	m.AddOperatingStage(operation.Joiner(joiners3))
	/* */

	m.AddOperatingStage(operation.Persist())
}

func (f *operationFactory) sarStages(scenario string, m *model.Model, _ int) []chan struct{} {
	joiners := make([]chan struct{}, 0)
	for _, host := range m.SelectHosts("*") {
		h := host // because stage is func (closure)
		joiner := make(chan struct{})
		m.AddOperatingStage(operation.Sar(scenario, h, 1, joiner))
		joiners = append(joiners, joiner)
	}
	return joiners
}

func (f *operationFactory) sarCloserStages(m *model.Model) {
	for _, host := range m.SelectHosts("*") {
		m.AddOperatingStage(operation.SarCloser(host))
	}
}

type operationFactory struct{}
