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
	"time"
)

func newStageFactory() model.Factory {
	return &stageFactory{}
}

func (self *stageFactory) Build(m *model.Model) error {
	m.MetricsHandlers = append(m.MetricsHandlers, model.StdOutMetricsWriter{})

	m.Infrastructure = model.InfrastructureStages{
		aws_ssh_keys0.Express(),
		terraform0.Express(),
		semaphore0.Restart(90 * time.Second),
	}

	m.Configuration = model.ConfigurationStages{
		zitilib_runlevel_1_configuration.IfNoPki(zitilib_runlevel_1_configuration.Fabric(), zitilib_runlevel_1_configuration.DotZiti()),
		config.Component(),
		config.Static([]config.StaticConfig{
			{Src: "loop/10-ambient.loop2.yml", Name: "10-ambient.loop2.yml"},
			{Src: "loop/4k-chatter.loop2.yml", Name: "4k-chatter.loop2.yml"},
		}),
		zitilib_bootstrap.DefaultZitiBinaries(),
	}

	m.Distribution = model.DistributionStages{
		distribution.Locations("*", "logs"),
		rsync.Parallel(),
	}

	m.AddActivationActions("bootstrap", "start")

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

	m.AddOperatingActions("syncModelEdgeState")
	m.AddOperatingStage(zitilib_5_operation.Mesh(runPhase.GetCloser()))
	m.AddOperatingStage(zitilib_5_operation.ModelMetricsWithIdMapper(runPhase.GetCloser(), func(id string) string {
		return "component.edgeId:" + id
	}))

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
		remoteConfigFile := "/home/fedora/fablab/cfg/" + c.PublicIdentity + ".json"
		stage := zitilib_5_operation.LoopListener(c.GetHost(), nil, "edge:perf-test", "--config-file", remoteConfigFile)
		m.AddOperatingStage(stage)
	}

	return nil
}

func (_ *stageFactory) dialers(m *model.Model, phase fablib_5_operation.Phase) error {
	components := m.SelectComponents(models.ClientTag)
	if len(components) < 1 {
		return fmt.Errorf("no '%v' components in model", models.ClientTag)
	}

	for _, c := range components {
		remoteConfigFile := "/home/fedora/fablab/cfg/" + c.PublicIdentity + ".json"
		stage := zitilib_5_operation.LoopDialer(c.GetHost(), "10-ambient.loop2.yml", "edge:perf-test", phase.AddJoiner(), "--config-file", remoteConfigFile)
		m.AddOperatingStage(stage)
	}

	return nil
}

type stageFactory struct{}
