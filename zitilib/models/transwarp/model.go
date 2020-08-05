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
				"ctrl": {
					Components: model.Components{
						"ctrl": {
							BinaryName:     "ziti-controller",
							ConfigSrc:      "ctrl.yml",
							ConfigName:     "ctrl.yml",
							PublicIdentity: "ctrl",
							Scope:          model.Scope{Tags: model.Tags{"ctrl"}},
						},
					},
					Scope: model.Scope{Tags: model.Tags{"ctrl"}},
				},
				"router": {
					Components: model.Components{
						"local": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "egress_router.yml",
							ConfigName:     "router.yml",
							PublicIdentity: "local",
							Scope:          model.Scope{Tags: model.Tags{"router", "terminator"}},
						},
					},
					Scope: model.Scope{Tags: model.Tags{"router", "terminator"}},
				},
				"service": {
					Scope: model.Scope{Tags: model.Tags{"service", "iperf_server"}},
				},
			},
			Id:    "us-east-1",
			Az:    "us-east-1c",
			Scope: model.Scope{Tags: model.Tags{"ctrl", "router", "terminator", "iperf_server"}},
		},
		"remote": {
			Hosts: model.Hosts{
				"router": {
					Components: model.Components{
						"remote": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "ingress_router.yml",
							ConfigName:     "remote.yml",
							PublicIdentity: "remote",
							Scope:          model.Scope{Tags: model.Tags{"router", "initiator"}},
						},
					},
					Scope: model.Scope{Tags: model.Tags{"router", "initiator"}},
				},
				"client": {
					Scope: model.Scope{Tags: model.Tags{"client", "iperf_client"}},
				},
			},
			Id:    "us-west-1",
			Az:    "us-west-1a",
			Scope: model.Scope{Tags: model.Tags{"client", "router", "initiator", "iperf_client"}},
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
}
