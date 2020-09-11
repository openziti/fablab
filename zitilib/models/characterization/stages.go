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
	aws_ssh_keys0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/aws_ssh_key"
	semaphore0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/semaphore"
	terraform0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/terraform"
	"github.com/openziti/fablab/kernel/fablib/runlevel/1_configuration/config"
	distribution "github.com/openziti/fablab/kernel/fablib/runlevel/3_distribution"
	"github.com/openziti/fablab/kernel/fablib/runlevel/3_distribution/rsync"
	operation "github.com/openziti/fablab/kernel/fablib/runlevel/5_operation"
	aws_ssh_keys6 "github.com/openziti/fablab/kernel/fablib/runlevel/6_disposal/aws_ssh_key"
	terraform6 "github.com/openziti/fablab/kernel/fablib/runlevel/6_disposal/terraform"
	"github.com/openziti/fablab/kernel/model"
	zitilib_bootstrap "github.com/openziti/fablab/zitilib"
	"github.com/openziti/fablab/zitilib/models"
	zitilib_runlevel_1_configuration "github.com/openziti/fablab/zitilib/runlevel/1_configuration"
	zitilib_runlevel_5_operation "github.com/openziti/fablab/zitilib/runlevel/5_operation"
	"time"
)

func newStagesFactory() model.Factory {
	return &stagesFactory{}
}

func (f *stagesFactory) Build(m *model.Model) error {
	m.Infrastructure = model.InfrastructureStages{
		aws_ssh_keys0.Express(),
		terraform0.Express(),
		semaphore0.Restart(90 * time.Second),
	}

	m.Configuration = model.ConfigurationStages{
		zitilib_runlevel_1_configuration.IfNoPki(zitilib_runlevel_1_configuration.Fabric(), zitilib_runlevel_1_configuration.DotZiti()),
		config.Component(),
		zitilib_bootstrap.DefaultZitiBinaries(),
	}

	m.Distribution = model.DistributionStages{
		distribution.Locations(models.ControllerTag, "logs"),
		distribution.Locations(models.RouterTag, "logs"),
		rsync.Rsync(),
	}

	m.AddActivationActions("bootstrap", "start")

	if err := f.addOperationStages(m); err != nil {
		return err
	}

	m.Disposal = model.DisposalStages{
		terraform6.Dispose(),
		aws_ssh_keys6.Dispose(),
	}

	return nil
}

func (f *stagesFactory) addOperationStages(m *model.Model) error {
	value, err := m.SelectHost("#local > #service")
	if err != nil {
		return err
	}
	directEndpoint := value.PublicIp

	if value, err = m.SelectHost("#short > #short"); err != nil {
		return err
	}
	shortProxy := value.PrivateIp

	if value, err = m.SelectHost("#medium > #medium"); err != nil {
		return err
	}
	mediumProxy := value.PrivateIp

	if value, err = m.SelectHost("#long > #long"); err != nil {
		return err
	}
	longProxy := value.PrivateIp

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

func (f *stagesFactory) addStagesForRegion(region, initiatingRouter, directEndpoint string, tcpdump bool, snaplen, seconds int, m *model.Model) {
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

func (f *stagesFactory) sarStages(scenario string, m *model.Model, _ int) []chan struct{} {
	joiners := make([]chan struct{}, 0)
	for _, host := range m.SelectHosts("*") {
		h := host // because stage is func (closure)
		joiner := make(chan struct{})
		m.AddOperatingStage(operation.Sar(scenario, h, 1, joiner))
		joiners = append(joiners, joiner)
	}
	return joiners
}

func (f *stagesFactory) sarCloserStages(m *model.Model) {
	for _, host := range m.SelectHosts("*") {
		m.AddOperatingStage(operation.SarCloser(host))
	}
}

type stagesFactory struct{}
