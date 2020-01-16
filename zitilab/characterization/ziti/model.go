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

package zitilab_characterization_ziti

import "github.com/netfoundry/fablab/kernel/model"

func init() {
	model.RegisterModel("zitilab/characterization/ziti", Model)
}

// Static model skeleton for zitilab/characterization/ziti
//
var Model = &model.Model{
	Factories: []model.Factory{
		newModelScopeFactory(),
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
		"local": {
			Scope: model.Scope{Tags: model.Tags{"ctrl", "router", "iperf_server"}},
			Id:    "us-east-1",
			Az:    "us-east-1a",
			Hosts: model.Hosts{
				"ctrl": {
					Scope: model.Scope{Tags: model.Tags{"ctrl"}},
					Components: model.Components{
						"ctrl": {
							Scope:          model.Scope{Tags: model.Tags{"ctrl"}},
							BinaryName:     "ziti-controller",
							ConfigSrc:      "ctrl.yml",
							ConfigName:     "ctrl.yml",
							PublicIdentity: "ctrl",
						},
					},
				},
				"local": {
					Scope: model.Scope{Tags: model.Tags{"router", "terminator"}},
					Components: model.Components{
						"local": {
							Scope:          model.Scope{Tags: model.Tags{"router", "terminator"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "egress_router.yml",
							ConfigName:     "local.yml",
							PublicIdentity: "local",
						},
					},
				},
				"service": {
					Scope: model.Scope{Tags: model.Tags{"service", "iperf_server"}},
				},
			},
		},
		"short": {
			Scope: model.Scope{Tags: model.Tags{"router"}},
			Id:    "us-west-1",
			Az:    "us-west-1c",
			Hosts: model.Hosts{
				"short": {
					Scope: model.Scope{Tags: model.Tags{"router"}},
					Components: model.Components{
						"short": {
							Scope:          model.Scope{Tags: model.Tags{"router"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "ingress_router.yml",
							ConfigName:     "short.yml",
							PublicIdentity: "short",
						},
					},
				},
				"client": {
					Scope: model.Scope{Tags: model.Tags{"client", "iperf_client"}},
				},
			},
		},
		"medium": {
			Scope: model.Scope{Tags: model.Tags{"router"}},
			Id:    "ap-south-1",
			Az:    "ap-south-1a",
			Hosts: model.Hosts{
				"medium": {
					Scope: model.Scope{Tags: model.Tags{"router"}},
					Components: model.Components{
						"medium": {
							Scope:          model.Scope{Tags: model.Tags{"router"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "ingress_router.yml",
							ConfigName:     "medium.yml",
							PublicIdentity: "medium",
						},
					},
				},
				"client": {
					Scope: model.Scope{Tags: model.Tags{"client", "iperf_client"}},
				},
			},
		},
		"long": {
			Scope: model.Scope{Tags: model.Tags{"router"}},
			Id:    "ap-southeast-2",
			Az:    "ap-southeast-2c",
			Hosts: model.Hosts{
				"long": {
					Scope: model.Scope{Tags: model.Tags{"router"}},
					Components: model.Components{
						"long": {
							Scope:          model.Scope{Tags: model.Tags{"router"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "ingress_router.yml",
							ConfigName:     "long.yml",
							PublicIdentity: "long",
						},
					},
				},
				"client": {
					Scope: model.Scope{Tags: model.Tags{"client", "iperf_client"}},
				},
			},
		},
	},
}
