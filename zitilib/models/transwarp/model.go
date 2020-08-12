/*
	Copyright NetFoundry, Inc.

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

package transwarp

import (
	"github.com/openziti/fablab/kernel/model"
	zitilib_transwarp_actions "github.com/openziti/fablab/zitilib/models/transwarp/actions"
)

func init() {
	model.RegisterModel("zitilib/transwarp", transwarpModel)
}

var transwarpModel = &model.Model{
	Regions: model.Regions{
		"local": {
			Hosts: model.Hosts{
				"local": {
					Components: model.Components{
						"ctrl": {
							BinaryName:     "ziti-controller",
							ConfigSrc:      "ctrl.yml",
							ConfigName:     "ctrl.yml",
							PublicIdentity: "ctrl",
							Scope:          model.Scope{Tags: model.Tags{"ctrl"}},
						},
						"local": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "transwarp_ingress_router.yml",
							ConfigName:     "local.yml",
							PublicIdentity: "local",
							Scope:          model.Scope{Tags: model.Tags{"router", "initiator"}},
						},
					},
					InstanceType: "t2.medium",
					Scope:        model.Scope{Tags: model.Tags{"ctrl", "router", "initiator", "iperf_client"}},
				},
			},
			Id:    "us-east-1",
			Az:    "us-east-1c",
			Scope: model.Scope{Tags: model.Tags{"ctrl", "router", "initiator", "iperf_client"}},
		},
		"remote": {
			Hosts: model.Hosts{
				"remote": {
					InstanceType: "t2.medium",
					Components: model.Components{
						"remote": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "transwarp_egress_router.yml",
							ConfigName:     "remote.yml",
							PublicIdentity: "remote",
							Scope:          model.Scope{Tags: model.Tags{"router", "terminator"}},
						},
					},
					Scope: model.Scope{Tags: model.Tags{"router", "terminator", "client", "iperf_server"}},
				},
			},
			Id:    "ap-southeast-2",
			Az:    "ap-southeast-2c",
			Scope: model.Scope{Tags: model.Tags{"client", "router", "terminator", "iperf_server"}},
		},
	},

	Scope: model.Scope{
		Variables: model.Variables{
			"zitilib": model.Variables{
				"fabric": model.Variables{
					"data_plane_protocol": &model.Variable{Default: "transwarp"},
				},
			},
			"characterization": model.Variables{
				"sample_minutes": &model.Variable{Default: 1},
				"tcpdump": model.Variables{
					"enabled": &model.Variable{Default: true},
					"snaplen": &model.Variable{Default: 128},
				},
			},
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
			"distribution": model.Variables{
				"rsync_bin": &model.Variable{Default: "rsync"},
				"ssh_bin":   &model.Variable{Default: "ssh"},
			},
		},
	},

	Factories: []model.Factory{
		newHostsFactory(),
		newInfrastructureFactory(),
		newConfigurationFactory(),
		newKittingFactory(),
		newDistributionFactory(),
		newActivationFactory(),
		newOperationFactory(),

		zitilib_transwarp_actions.NewActionsFactory(),
	},

	BootstrapExtensions: []model.BootstrapExtension{
		&bootstrap{},
	},
}
