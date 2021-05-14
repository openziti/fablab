package edge

import (
	"fmt"
	aws_ssh_keys0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/aws_ssh_key"
	semaphore0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/semaphore"
	terraform0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/terraform"
	"github.com/openziti/fablab/kernel/fablib/runlevel/1_configuration/config"
	distribution "github.com/openziti/fablab/kernel/fablib/runlevel/3_distribution"
	"github.com/openziti/fablab/kernel/fablib/runlevel/3_distribution/rsync"
	fablib_5_operation "github.com/openziti/fablab/kernel/fablib/runlevel/5_operation"
	aws_ssh_keys6 "github.com/openziti/fablab/kernel/fablib/runlevel/6_disposal/aws_ssh_key"
	terraform6 "github.com/openziti/fablab/kernel/fablib/runlevel/6_disposal/terraform"
	"github.com/openziti/fablab/kernel/model"
	zitilib_bootstrap "github.com/openziti/fablab/zitilib"
	"github.com/openziti/fablab/zitilib/models"
	zitilib_runlevel_1_configuration "github.com/openziti/fablab/zitilib/runlevel/1_configuration"
	zitilib_5_operation "github.com/openziti/fablab/zitilib/runlevel/5_operation"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

func newStageFactory() model.Factory {
	return &stageFactory{}
}

func (self *stageFactory) Build(m *model.Model) error {
	// m.MetricsHandlers = append(m.MetricsHandlers, model.StdOutMetricsWriter{})

	m.Infrastructure = model.InfrastructureStages{
		aws_ssh_keys0.Express(),
		terraform0.Express(),
		semaphore0.Restart(90 * time.Second),
	}

	m.Configuration = model.ConfigurationStages{
		zitilib_runlevel_1_configuration.IfNoPki(zitilib_runlevel_1_configuration.Fabric(), zitilib_runlevel_1_configuration.DotZiti()),
		config.Component(),
		config.Static([]config.StaticConfig{
			{Src: "remote_identities.yml", Name: "remote_identities.yml"},
		}),
		zitilib_bootstrap.DefaultZitiBinaries(),
	}

	m.Distribution = model.DistributionStages{
		distribution.DistributeSshKey("*"),
		distribution.Locations("*", "logs"),
		rsync.Rsync(25),
	}

	m.AddActivationActions("stop", "bootstrap", "start")

	if err := self.addOperationStages(m); err != nil {
		return err
	}

	m.Disposal = model.DisposalStages{
		terraform6.Dispose(),
		aws_ssh_keys6.Dispose(),
	}

	return nil
}

func (self *stageFactory) addOperationStages(m *model.Model) error {
	runPhase := fablib_5_operation.NewPhase()
	cleanupPhase := fablib_5_operation.NewPhase()

	clientMetrics := zitilib_5_operation.NewClientMetrics("metrics", runPhase.GetCloser())
	m.AddActivationStage(clientMetrics)

	m.AddOperatingActions("syncModelEdgeState")
	m.AddOperatingStage(fablib_5_operation.InfluxMetricsReporter())
	m.AddOperatingStage(zitilib_5_operation.Mesh(runPhase.GetCloser()))
	m.AddOperatingStage(zitilib_5_operation.ModelMetricsWithIdMapper(runPhase.GetCloser(), func(id string) string {
		if id == "ctrl" {
			return "#ctrl"
		}
		id = strings.ReplaceAll(id, ".", ":")
		return "component.edgeId:" + id
	}))
	m.AddOperatingStage(clientMetrics)

	for _, host := range m.SelectHosts("*") {
		m.AddOperatingStage(fablib_5_operation.StreamSarMetrics(host, 5, 3, runPhase, cleanupPhase))
	}

	if err := self.listeners(m); err != nil {
		return fmt.Errorf("error creating listeners (%w)", err)
	}

	m.AddOperatingStage(fablib_5_operation.Timer(5*time.Second, nil))

	if err := self.dialers(m, runPhase); err != nil {
		return fmt.Errorf("error creating dialers (%w)", err)
	}

	m.AddOperatingStage(runPhase)
	m.AddOperatingStage(fablib_5_operation.Persist())

	return nil
}

func (_ *stageFactory) listeners(m *model.Model) error {
	components := m.SelectComponents(models.ServiceTag)
	if len(components) < 1 {
		return fmt.Errorf("no '%v' components in model", models.ServiceTag)
	}

	for _, c := range components {
		// remoteConfigFile := fmt.Sprintf("/home/%v/fablab/cfg/%v.json", m.MustVariable("credentials", "ssh", "username"), c.PublicIdentity)
		//stage := zitilib_5_operation.Loop3Listener(c.GetHost(), nil, "edge:perf-test", "--config-file", remoteConfigFile)
		stage := zitilib_5_operation.Loop3Listener(c.GetHost(), nil, "tcp:0.0.0.0:8171")
		m.AddOperatingStage(stage)
	}

	return nil
}

func (_ *stageFactory) dialers(m *model.Model, phase fablib_5_operation.Phase) error {
	var components []*model.Component
	components = m.SelectComponents(models.ClientTag)
	//for i := 0; i < 26; i++ {
	//	components = append(components, m.SelectComponents(fmt.Sprintf("#client%v", i))...)
	//}
	if len(components) < 1 {
		return fmt.Errorf("no '%v' components in model", models.ClientTag)
	}

	initiator, err := m.SelectHost(".initiator")
	if err != nil {
		return err
	}
	log.Debug("initiator: %v", initiator.PublicIp)

	for _, c := range components {
		remoteConfigFile := fmt.Sprintf("/home/%v/fablab/cfg/%v.json", m.MustVariable("credentials", "ssh", "username"), c.PublicIdentity)
		loopFile := fmt.Sprintf("edge-perf-%v.loop3.yml", c.Host.Index)
		stage := zitilib_5_operation.Loop3Dialer(c.GetHost(), loopFile, "edge:perf-test", phase.AddJoiner(), "--config-file", remoteConfigFile)
		//endpoint := fmt.Sprintf("tcp:%v:7001", initiator.PublicIp)
		//stage := zitilib_5_operation.Loop3Dialer(c.GetHost(), loopFile, endpoint, phase.AddJoiner(), "--config-file", remoteConfigFile, "--direct")
		m.AddOperatingStage(stage)
	}

	return nil
}

type stageFactory struct{}
