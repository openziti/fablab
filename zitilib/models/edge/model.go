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

package edge

import (
	"github.com/openziti/fablab/kernel/model"
)

func init() {
	model.RegisterModel("zitilib/edge", edge)
}

// Static model skeleton for zitilib/edge
//
var edge = &model.Model{
	Scope: model.Scope{
		Variables: model.Variables{
			"edge": model.Variables{
				"region": &model.Variable{Default: "us-east-1"},
				"az":     &model.Variable{Default: "us-east-1c"},
				"sizing": model.Variables{
					"ctrl":       &model.Variable{Default: "t2.medium"},
					"initiator":  &model.Variable{Default: "t2.medium"},
					"terminator": &model.Variable{Default: "t2.medium"},
					"client":     &model.Variable{Default: "t2.medium"},
					"service":    &model.Variable{Default: "t2.medium"},
				},
			},
			"zitilib": model.Variables{ //configs in /lib link to this....
				"fabric": model.Variables{
					"data_plane_protocol": &model.Variable{Default: "tls"},
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
				"edge": model.Variables{
					"username": &model.Variable{Required: true, Sensitive: true},
					"password": &model.Variable{Required: true, Sensitive: true},
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
			Scope:  model.Scope{Tags: model.Tags{"initiator"}},
			Region: "us-east-1",
			Site:   "us-east-1a",
			Hosts: model.Hosts{
				"ctrl": {
					Scope: model.Scope{Tags: model.Tags{"ctrl"}},
					Components: model.Components{
						"ctrl": {
							Scope:          model.Scope{Tags: model.Tags{"ctrl"}},
							BinaryName:     "ziti-controller",
							ConfigSrc:      "ctrl_edge.yml",
							ConfigName:     "ctrl_edge.yml",
							PublicIdentity: "ctrl",
						},
					},
				},
				"initiator": {
					Scope: model.Scope{Tags: model.Tags{"edge-router"}},
					Components: model.Components{
						"initiator": {
							Scope:          model.Scope{Tags: model.Tags{"edge-router"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "edge_router.yml",
							ConfigName:     "edge_router_initiator.yml",
							PublicIdentity: "edge_router_initiator",
						},
					},
				},
				"client": {
					Scope: model.Scope{Tags: model.Tags{"client", "sdk-app"}},
					Components: model.Components{
						"client1": {
							BinaryName:     "ziti-fabric-test",
							PublicIdentity: "client1",
						},
					},
				},
			},
		},
		"terminator": {
			Region: "us-west-1",
			Site:   "us-west-1b",
			Scope:  model.Scope{Tags: model.Tags{"terminator"}},
			Hosts: model.Hosts{
				"terminator": {
					Scope: model.Scope{Tags: model.Tags{"edge-router"}},
					Components: model.Components{
						"terminator": {
							Scope:          model.Scope{Tags: model.Tags{"edge-router"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "edge_router.yml",
							ConfigName:     "edge_router_terminator.yml",
							PublicIdentity: "edge_router_terminator",
						},
					},
				},
				"service": {
					Scope: model.Scope{Tags: model.Tags{"service", "sdk-app"}},
					Components: model.Components{
						"server1": {
							BinaryName:     "ziti-fabric-test",
							PublicIdentity: "server1",
						},
					},
				},
			},
		},
	},
}
