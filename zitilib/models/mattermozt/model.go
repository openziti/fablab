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

package mattermozt

import "github.com/netfoundry/fablab/kernel/model"

func init() {
	model.RegisterModel("zitilib/mattermozt", mattermozt)
}

// Static model skeleton for zitilib/mattermozt
//
var mattermozt = &model.Model{
	Scope: model.Scope{
		Variables: model.Variables{
			"mattermozt": model.Variables{
				"region": &model.Variable{Default: "us-east-1"},
				"az":     &model.Variable{Default: "us-east-1c"},
				"sizing": model.Variables{
					"ctrl":       &model.Variable{Default: "t2.medium"},
					"terminator": &model.Variable{Default: "t2.medium"},
					"edge":       &model.Variable{Default: "t2.medium"},
					"service":    &model.Variable{Default: "t2.medium"},
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
		newRegionFactory(),
		newHostsFactory(),
		newInfrastructureFactory(),
		newConfigurationFactory(),
		newKittingFactory(),
		newDistributionFactory(),
		// activation
		// operation
	},

	Regions: model.Regions{
		"local": {
			Scope: model.Scope{Tags: model.Tags{"ctrl", "router", "service"}},
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
				"terminator": {
					Scope: model.Scope{Tags: model.Tags{"router"}},
					Components: model.Components{
						"terminator": {
							Scope:          model.Scope{Tags: model.Tags{"router", "terminator"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "egress_router.yml",
							ConfigName:     "local.yml",
							PublicIdentity: "local",
						},
					},
				},
				"edge": {
					Scope: model.Scope{Tags: model.Tags{"edge", "router"}},
				},
				"service": {
					Scope: model.Scope{Tags: model.Tags{"service"}},
				},
			},
		},
	},
}
