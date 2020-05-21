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

package zitilib_examples

import (
	"github.com/openziti/fablab/kernel/model"
)

func init() {
	model.RegisterModel("zitilib/examples/smartrouting", smartrouting)
}

var smartrouting = &model.Model{
	Scope: modelScope,

	Factories: []model.Factory{
		newHostsFactory(),
		newActionsFactory(),
		newInfrastructureFactory(),
		newConfigurationFactory(),
		newKittingFactory(),
		newDistributionFactory(),
		newActivationFactory(),
		newOperationFactory(),
	},

	Regions: model.Regions{
		"initiator": {
			Scope: model.Scope{
				Tags: model.Tags{"ctrl", "router", "loop", "initiator", "iperf_server"},
			},
			Id: "us-east-1",
			Az: "us-east-1a",
			Hosts: model.Hosts{
				"ctrl": {
					Scope: model.Scope{Tags: model.Tags{"ctrl"}},
					Components: model.Components{
						"ctrl": {
							Scope: model.Scope{
								Tags: model.Tags{"ctrl"},
							},
							BinaryName:     "ziti-controller",
							ConfigSrc:      "ctrl.yml",
							ConfigName:     "ctrl.yml",
							PublicIdentity: "ctrl",
						},
					},
				},
				"001": {
					Scope: model.Scope{Tags: model.Tags{"router", "initiator", "iperf_server"}},
					Components: model.Components{
						"001": {
							Scope:          model.Scope{Tags: model.Tags{"router"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "ingress_router.yml",
							ConfigName:     "001.yml",
							PublicIdentity: "001",
						},
					},
				},
				"loop0": {
					Scope: model.Scope{Tags: model.Tags{"loop-dialer"}},
				},
				"loop1": {
					Scope: model.Scope{Tags: model.Tags{"loop-dialer"}},
				},
				"loop2": {
					Scope: model.Scope{Tags: model.Tags{"loop-dialer"}},
				},
				"loop3": {
					Scope: model.Scope{Tags: model.Tags{"loop-dialer"}},
				},
			},
		},
		"transitA": {
			Scope: model.Scope{Tags: model.Tags{"router"}},
			Id:    "us-west-1",
			Az:    "us-west-1b",
			Hosts: model.Hosts{
				"002": {
					Scope: model.Scope{Tags: model.Tags{"router"}},
					Components: model.Components{
						"002": {
							Scope:          model.Scope{Tags: model.Tags{"router"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "transit_router.yml",
							ConfigName:     "002.yml",
							PublicIdentity: "002",
						},
					},
				},
			},
		},
		"transitB": {
			Scope: model.Scope{Tags: model.Tags{"router"}},
			Id:    "us-east-2",
			Az:    "us-east-2c",
			Hosts: model.Hosts{
				"004": {
					Scope: model.Scope{Tags: model.Tags{"router"}},
					Components: model.Components{
						"004": {
							Scope:          model.Scope{Tags: model.Tags{"router"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "transit_router.yml",
							ConfigName:     "004.yml",
							PublicIdentity: "004",
						},
					},
				},
			},
		},
		"terminator": {
			Scope: model.Scope{Tags: model.Tags{"router", "loop", "terminator"}},
			Id:    "us-west-2",
			Az:    "us-west-2b",
			Hosts: model.Hosts{
				"003": {
					Scope: model.Scope{Tags: model.Tags{"router"}},
					Components: model.Components{
						"003": {
							Scope:          model.Scope{Tags: model.Tags{"router", "terminator"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "egress_router.yml",
							ConfigName:     "003.yml",
							PublicIdentity: "003",
						},
					},
				},
				"loop0": {
					Scope: model.Scope{Tags: model.Tags{"loop-listener"}},
				},
				"loop1": {
					Scope: model.Scope{Tags: model.Tags{"loop-listener"}},
				},
				"loop2": {
					Scope: model.Scope{Tags: model.Tags{"loop-listener"}},
				},
				"loop3": {
					Scope: model.Scope{Tags: model.Tags{"loop-listener"}},
				},
			},
		},
	},
}
