package models

import (
	"fmt"
	"github.com/netfoundry/fablab/actions"
	"github.com/netfoundry/fablab/actions/cli"
	"github.com/netfoundry/fablab/actions/component"
	"github.com/netfoundry/fablab/actions/host"
	"github.com/netfoundry/fablab/actions/metrics"
	"github.com/netfoundry/fablab/actions/semaphore"
	"github.com/netfoundry/fablab/kernel"
	semaphore0 "github.com/netfoundry/fablab/stages/0_infrastructure/semaphore"
	terraform0 "github.com/netfoundry/fablab/stages/0_infrastructure/terraform"
	"github.com/netfoundry/fablab/stages/1_configuration/config"
	"github.com/netfoundry/fablab/stages/1_configuration/pki"
	"github.com/netfoundry/fablab/stages/2_kitting/devkit"
	"github.com/netfoundry/fablab/stages/3_distribution/rsync"
	"github.com/netfoundry/fablab/stages/4_activation/action"
	operation "github.com/netfoundry/fablab/stages/5_operation"
	terraform6 "github.com/netfoundry/fablab/stages/6_disposal/terraform"
	"github.com/netfoundry/fablab/zitilab"
	zitilab_5_operation "github.com/netfoundry/fablab/zitilab/stages/5_operation"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

func init() {
	kernel.RegisterModel("diamondback", diamondback)
	kernel.RegisterModel("tiny", tiny)
	kernel.RegisterModel("transit", transit)
}

func commonActions() kernel.ActionBinders {
	return kernel.ActionBinders{
		"bootstrap": doBootstrap,
		"start":     doStart,
		"stop":      doStop,
		"metrics":   doMetrics,
	}
}

func commonInfrastructure() kernel.InfrastructureBinders {
	return kernel.InfrastructureBinders{
		func(m *kernel.Model) kernel.InfrastructureStage { return terraform0.Express() },
		func(m *kernel.Model) kernel.InfrastructureStage { return semaphore0.Restart(90 * time.Second) },
	}
}

func commonConfiguration() kernel.ConfigurationBinders {
	return kernel.ConfigurationBinders{
		func(m *kernel.Model) kernel.ConfigurationStage { return pki.Group(pki.Fabric(), pki.DotZiti()) },
		func(m *kernel.Model) kernel.ConfigurationStage { return config.Component() },
		func(m *kernel.Model) kernel.ConfigurationStage {
			configs := []config.StaticConfig{
				{Src: "loop/10-ambient.loop2.yml", Name: "10-ambient.loop2.yml"},
				{Src: "loop/4k-chatter.loop2.yml", Name: "4k-chatter.loop2.yml"},
				{Src: "remote_identities.yml", Name: "remote_identities.yml"},
			}
			return config.Static(configs)
		},
	}
}

func commonKitting() kernel.KittingBinders {
	return kernel.KittingBinders{
		func(m *kernel.Model) kernel.KittingStage {
			zitiBinaries := []string{
				"ziti-controller",
				"ziti-fabric",
				"ziti-fabric-test",
				"ziti-router",
			}
			return devkit.DevKit(filepath.Join(zitilab.ZitiRoot(), "bin"), zitiBinaries)
		},
	}
}

func commonDistribution() kernel.DistributionBinders {
	return kernel.DistributionBinders{
		func(m *kernel.Model) kernel.DistributionStage { return rsync.Rsync() },
	}
}

func commonActivation() kernel.ActivationBinders {
	return kernel.ActivationBinders{
		func(m *kernel.Model) kernel.ActivationStage { return action.Activation("bootstrap", "start") },
	}
}

func commonOperation() kernel.OperatingBinders {
	logrus.Infof("binding")
	c := make(chan struct{})
	return kernel.OperatingBinders{
		func(m *kernel.Model) kernel.OperatingStage { return zitilab_5_operation.Metrics(c) },
		func(m *kernel.Model) kernel.OperatingStage {
			minutes, found := m.GetVariable("sample_minutes")
			if !found {
				minutes = 1
			}
			return operation.Timer(time.Duration(minutes.(int))*time.Minute, c)
		},
		func(m *kernel.Model) kernel.OperatingStage { return operation.Persist() },
	}
}

func commonDisposal() kernel.DisposalBinders {
	return kernel.DisposalBinders{
		func(m *kernel.Model) kernel.DisposalStage { return terraform6.Dispose() },
	}
}

func doBootstrap(m *kernel.Model) kernel.Action {
	sshUsername := m.MustVariable("credentials", "ssh", "username").(string)

	workflow := actions.Workflow()

	workflow.AddAction(component.Stop("@ctrl", "@ctrl", "@ctrl"))
	workflow.AddAction(component.Start("@ctrl", "@ctrl", "@ctrl"))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))

	for _, router := range m.GetComponentsByTag("router") {
		cert := fmt.Sprintf("/intermediate/certs/%s-client.cert", router.PublicIdentity)
		workflow.AddAction(cli.Fabric("create", "router", filepath.Join(kernel.PkiBuild(), cert)))
	}

	components := m.GetComponentsByTag("terminator")
	serviceActions, err := createServiceActions(m, components[0].PublicIdentity)
	if err != nil {
		logrus.Fatalf("error creating service actions (%w)", err)
	}
	for _, serviceAction := range serviceActions {
		workflow.AddAction(serviceAction)
	}

	for _, h := range m.GetAllHosts() {
		workflow.AddAction(host.Exec(h, fmt.Sprintf("mkdir -p /home/%s/.ziti", sshUsername)))
		workflow.AddAction(host.Exec(h, fmt.Sprintf("rm -f /home/%s/.ziti/identities.yml", sshUsername)))
		workflow.AddAction(host.Exec(h, fmt.Sprintf("ln -s /home/%s/fablab/cfg/remote_identities.yml /home/%s/.ziti/identities.yml", sshUsername, sshUsername)))
	}

	workflow.AddAction(component.Stop("@ctrl", "@ctrl", "@ctrl"))

	return workflow
}

func doStart(m *kernel.Model) kernel.Action {
	sshUsername := m.MustVariable("credentials", "ssh", "username").(string)

	listenerCmd := fmt.Sprintf("nohup /home/%s/fablab/bin/ziti-fabric-test loop2 listener -b tcp:0.0.0.0:8171 > /home/%s/ziti-fabric-test.log 2>&1 &", sshUsername, sshUsername)

	workflow := actions.Workflow()
	workflow.AddAction(component.Start("@ctrl", "@ctrl", "@ctrl"))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))
	workflow.AddAction(component.Start("@router", "@router", "@router"))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))
	workflow.AddAction(host.GroupExec("@loop", "@loop-listener", listenerCmd))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))

	r001 := m.GetHosts("@initiator", "@initiator")
	if len(r001) != 1 {
		logrus.Fatalf("expected to find a single host tagged [initiator/initiator]")
	}
	endpoint := fmt.Sprintf("tls:%s:7001", r001[0].PublicIp)
	dialerActions, err := createDialerActions(m, endpoint)
	if err != nil {
		logrus.Fatalf("error creating dialer actions (%w)", err)
	}
	for _, dialerAction := range dialerActions {
		workflow.AddAction(dialerAction)
	}

	return workflow
}

func doStop(_ *kernel.Model) kernel.Action {
	return actions.Workflow(
		host.GroupKill("@loop", "@loop-dialer", "ziti-fabric-test"),
		host.GroupKill("@loop", "@loop-listener", "ziti-fabric-test"),
		component.Stop("@router", "@router", "@router"),
		component.Stop("@ctrl", "@ctrl", "@ctrl"),
	)
}

func doMetrics(_ *kernel.Model) kernel.Action {
	return metrics.Metrics()
}

func loopScenario(m *kernel.Model) string {
	loopScenario := "10-ambient.loop2.yml"
	if initiator := m.GetRegionByTag("initiator"); initiator != nil {
		if len(initiator.Hosts) > 1 {
			loopScenario = "4k-chatter.loop2.yml"
		}
	}
	return loopScenario
}

func createDialerActions(m *kernel.Model, endpoint string) ([]kernel.Action, error) {
	initiatorRegion := m.GetRegionByTag("initiator")
	if initiatorRegion == nil {
		return nil, fmt.Errorf("unable to find 'initiator' region")
	}

	sshUsername := m.MustVariable("credentials", "ssh", "username").(string)
	loopScenario := loopScenario(m)
	dialerActions := make([]kernel.Action, 0)
	for hostId, h := range initiatorRegion.Hosts {
		for _, tag := range h.Tags {
			if tag == "loop-dialer" {
				dialerCmd := fmt.Sprintf("nohup /home/%s/fablab/bin/ziti-fabric-test loop2 dialer /home/%s/fablab/cfg/%s -e %s -s %s > /home/%s/ziti-fabric-test.log 2>&1 &", sshUsername, sshUsername, loopScenario, endpoint, hostId, sshUsername)
				dialerActions = append(dialerActions, host.Exec(h, dialerCmd))
			}
		}
	}

	return dialerActions, nil
}

func createServiceActions(m *kernel.Model, terminatorId string) ([]kernel.Action, error) {
	terminatorRegion := m.GetRegionByTag("terminator")
	if terminatorRegion == nil {
		return nil, fmt.Errorf("unable to find 'terminator' region")
	}

	serviceActions := make([]kernel.Action, 0)
	for hostId, host := range terminatorRegion.Hosts {
		for _, tag := range host.Tags {
			if tag == "loop-listener" {
				serviceActions = append(serviceActions, cli.Fabric("create", "service", hostId, fmt.Sprintf("tcp:%s:8171", host.PrivateIp), terminatorId))
			}
		}
	}

	return serviceActions, nil
}

var kernelScope = kernel.Scope{
	Variables: kernel.Variables{
		"environment": &kernel.Variable{Required: true},
		"credentials": kernel.Variables{
			"aws": kernel.Variables{
				"access_key":   &kernel.Variable{Required: true},
				"secret_key":   &kernel.Variable{Required: true},
				"ssh_key_name": &kernel.Variable{Required: true},
			},
			"ssh": kernel.Variables{
				"key_path": &kernel.Variable{Required: true},
				"username": &kernel.Variable{Default: "fedora"},
			},
		},
		"sample_minutes": &kernel.Variable{Default: 1},
	},
}

var instanceType = func(def string) *kernel.Variable {
	return &kernel.Variable{
		Scoped:         true,
		GlobalFallback: true,
		Default:        def,
		Binder: func(v *kernel.Variable, i interface{}, path ...string) {
			if h, ok := i.(*kernel.Host); ok {
				h.InstanceType = v.Value.(string)
				logrus.Debugf("setting instance type of host %v = [%s]", path, h.InstanceType)
			}
		},
	}
}
