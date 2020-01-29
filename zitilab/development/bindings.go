/*
	Copyright 2019 NetFoundry, Inc.

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

package zitilab_development

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/actions"
	"github.com/netfoundry/fablab/kernel/actions/cli"
	"github.com/netfoundry/fablab/kernel/actions/component"
	"github.com/netfoundry/fablab/kernel/actions/host"
	"github.com/netfoundry/fablab/kernel/actions/semaphore"
	"github.com/netfoundry/fablab/kernel/model"
	semaphore0 "github.com/netfoundry/fablab/kernel/runlevel/0_infrastructure/semaphore"
	terraform0 "github.com/netfoundry/fablab/kernel/runlevel/0_infrastructure/terraform"
	"github.com/netfoundry/fablab/kernel/runlevel/1_configuration/config"
	"github.com/netfoundry/fablab/kernel/runlevel/1_configuration/pki"
	"github.com/netfoundry/fablab/kernel/runlevel/2_kitting/devkit"
	"github.com/netfoundry/fablab/kernel/runlevel/3_distribution/rsync"
	"github.com/netfoundry/fablab/kernel/runlevel/4_activation/action"
	operation "github.com/netfoundry/fablab/kernel/runlevel/5_operation"
	terraform6 "github.com/netfoundry/fablab/kernel/runlevel/6_disposal/terraform"
	"github.com/netfoundry/fablab/zitilab/characterization/reporting"
	zitilab_bootstrap "github.com/netfoundry/fablab/zitilab/development/bootstrap"
	"github.com/netfoundry/fablab/zitilab/development/console"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

func init() {
	model.RegisterModel("zitilab/development/diamondback", diamondback)
	model.RegisterModel("zitilab/development/tiny", tiny)
}

func commonActions() model.ActionBinders {
	return model.ActionBinders{
		"bootstrap": doBootstrap,
		"start":     doStart,
		"stop":      doStop,
		"console":   doConsole,
		"report":    doReport,
	}
}

func commonInfrastructure() model.InfrastructureBinders {
	return model.InfrastructureBinders{
		func(m *model.Model) model.InfrastructureStage { return terraform0.Express() },
		func(m *model.Model) model.InfrastructureStage { return semaphore0.Restart(90 * time.Second) },
	}
}

func commonConfiguration() model.ConfigurationBinders {
	return model.ConfigurationBinders{
		func(m *model.Model) model.ConfigurationStage { return pki.Group(pki.Fabric(), pki.DotZiti()) },
		func(m *model.Model) model.ConfigurationStage { return config.Component() },
		func(m *model.Model) model.ConfigurationStage {
			configs := []config.StaticConfig{
				{Src: "loop/10-ambient.loop2.yml", Name: "10-ambient.loop2.yml"},
				{Src: "loop/4k-chatter.loop2.yml", Name: "4k-chatter.loop2.yml"},
				{Src: "remote_identities.yml", Name: "remote_identities.yml"},
			}
			return config.Static(configs)
		},
	}
}

func commonKitting() model.KittingBinders {
	return model.KittingBinders{
		func(m *model.Model) model.KittingStage {
			zitiBinaries := []string{
				"ziti-controller",
				"ziti-fabric",
				"ziti-fabric-test",
				"ziti-router",
			}
			return devkit.DevKit(filepath.Join(zitilab_bootstrap.ZitiRoot(), "bin"), zitiBinaries)
		},
	}
}

func commonDistribution() model.DistributionBinders {
	return model.DistributionBinders{
		func(m *model.Model) model.DistributionStage { return rsync.Rsync() },
	}
}

func commonActivation() model.ActivationBinders {
	return model.ActivationBinders{
		func(m *model.Model) model.ActivationStage { return action.Activation("bootstrap", "start") },
	}
}

func commonOperation() model.OperatingBinders {
	c := make(chan struct{})
	binders := model.OperatingBinders{
		func(m *model.Model) model.OperatingStage { return operation.Mesh(c) },
		func(m *model.Model) model.OperatingStage { return operation.Metrics(c) },
		func(m *model.Model) model.OperatingStage {
			minutes, found := m.GetVariable("sample_minutes")
			if !found {
				minutes = 1
			}
			sampleDuration := time.Duration(minutes.(int)) * time.Minute

			values := m.GetHosts("@initiator", "@initiator")
			if len(values) == 1 {
				initiator := values[0].PublicIp
				return operation.Iperf(
					"ziti",
					initiator,
					"@iperf_server", "@iperf_server",
					"@iperf_client", "@iperf_client",
					int(sampleDuration.Seconds()),
				)
			}

			logrus.Fatalf("need single @initiator:@initiator host, found [%d]", len(values))
			return nil
		},
		func(m *model.Model) model.OperatingStage { return operation.Closer(c) },
		func(m *model.Model) model.OperatingStage { return operation.Persist() },
	}
	return binders
}

func commonDisposal() model.DisposalBinders {
	return model.DisposalBinders{
		func(m *model.Model) model.DisposalStage { return terraform6.Dispose() },
	}
}

func doBootstrap(m *model.Model) model.Action {
	sshUsername := m.MustVariable("credentials", "ssh", "username").(string)

	workflow := actions.Workflow()

	workflow.AddAction(component.Stop("@ctrl", "@ctrl", "@ctrl"))
	workflow.AddAction(component.Start("@ctrl", "@ctrl", "@ctrl"))
	workflow.AddAction(semaphore.Sleep(2 * time.Second))

	for _, router := range m.GetComponentsByTag("router") {
		cert := fmt.Sprintf("/intermediate/certs/%s-client.cert", router.PublicIdentity)
		workflow.AddAction(cli.Fabric("create", "router", filepath.Join(model.PkiBuild(), cert)))
	}

	iperfServer := m.GetHostByTags("iperf_server", "iperf_server")
	if iperfServer != nil {
		terminatingRouters := m.GetComponentsByTag("terminator")
		if len(terminatingRouters) < 1 {
			logrus.Fatal("need at least 1 terminating router!")
		}
		workflow.AddAction(cli.Fabric("create", "service", "iperf", "tcp:"+iperfServer.PublicIp+":7001", terminatingRouters[0].PublicIdentity))
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

func doStart(m *model.Model) model.Action {
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

func doStop(_ *model.Model) model.Action {
	return actions.Workflow(
		host.GroupKill("@loop", "@loop-dialer", "ziti-fabric-test"),
		host.GroupKill("@loop", "@loop-listener", "ziti-fabric-test"),
		component.Stop("@router", "@router", "@router"),
		component.Stop("@ctrl", "@ctrl", "@ctrl"),
	)
}

func doConsole(_ *model.Model) model.Action {
	return console.Console()
}

func doReport(_ *model.Model) model.Action {
	return reporting.Report()
}

func loopScenario(m *model.Model) string {
	loopScenario := "10-ambient.loop2.yml"
	if initiator := m.GetRegionByTag("initiator"); initiator != nil {
		if len(initiator.Hosts) > 1 {
			loopScenario = "4k-chatter.loop2.yml"
		}
	}
	return loopScenario
}

func createDialerActions(m *model.Model, endpoint string) ([]model.Action, error) {
	initiatorRegion := m.GetRegionByTag("initiator")
	if initiatorRegion == nil {
		return nil, fmt.Errorf("unable to find 'initiator' region")
	}

	sshUsername := m.MustVariable("credentials", "ssh", "username").(string)
	loopScenario := loopScenario(m)
	dialerActions := make([]model.Action, 0)
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

func createServiceActions(m *model.Model, terminatorId string) ([]model.Action, error) {
	terminatorRegion := m.GetRegionByTag("terminator")
	if terminatorRegion == nil {
		return nil, fmt.Errorf("unable to find 'terminator' region")
	}

	serviceActions := make([]model.Action, 0)
	for hostId, host := range terminatorRegion.Hosts {
		for _, tag := range host.Tags {
			if tag == "loop-listener" {
				serviceActions = append(serviceActions, cli.Fabric("create", "service", hostId, fmt.Sprintf("tcp:%s:8171", host.PrivateIp), terminatorId))
			}
		}
	}

	return serviceActions, nil
}

var kernelScope = model.Scope{
	Variables: model.Variables{
		"environment": &model.Variable{Required: true},
		"credentials": model.Variables{
			"aws": model.Variables{
				"access_key":   &model.Variable{Required: true, Sensitive: true},
				"secret_key":   &model.Variable{Required: true, Sensitive: true},
				"ssh_key_name": &model.Variable{Required: true},
			},
			"ssh": model.Variables{
				"key_path": &model.Variable{Required: true},
				"username": &model.Variable{Default: "fedora"},
			},
		},
		"sample_minutes": &model.Variable{Default: 1},
	},
}

var instanceType = func(def string) *model.Variable {
	return &model.Variable{
		Scoped:         true,
		GlobalFallback: true,
		Default:        def,
		Binder: func(v *model.Variable, i interface{}, path ...string) {
			if h, ok := i.(*model.Host); ok {
				h.InstanceType = v.Value.(string)
				logrus.Debugf("setting instance type of host %v = [%s]", path, h.InstanceType)
			}
		},
	}
}
