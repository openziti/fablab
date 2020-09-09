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
	"github.com/openziti/fablab/kernel/fablib/binding"
	"github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/aws_ssh_key"
	"github.com/openziti/fablab/kernel/model"
)

func init() {
	model.RegisterModel("zitilib/examples/tinyspot", tinyspot)
	model.AddBootstrapExtension(binding.AwsCredentialsLoader)
	model.AddBootstrapExtension(aws_ssh_key.KeyManager)
}

var tinyspot = &model.Model{

	Scope: modelScope,

	Factories: []model.Factory{
		newHostsFactory(),
		newActionsFactory(),
		newInfrastructureFactory(),
		newConfigurationFactory(),
		newDistributionFactory(),
		newActivationFactory(),
		newOperationFactory(),
	},

	Regions: model.Regions{
		"tiny": {
			Region: "us-east-1",
			Site:   "us-east-1c",
			Hosts: model.Hosts{
				"001": {
					InstanceResourceType: "spot",
					SpotPrice:            "0.02",
					Scope:                model.Scope{Tags: model.Tags{"loop-dialer", "loop-listener", "iperf_server"}},
					Components: model.Components{
						"ctrl": {
							Scope:          model.Scope{Tags: model.Tags{"ctrl"}},
							BinaryName:     "ziti-controller",
							ConfigSrc:      "ctrl.yml",
							ConfigName:     "ctrl.yml",
							PublicIdentity: "ctrl",
						},
						"001": {
							Scope:          model.Scope{Tags: model.Tags{"router", "initiator", "terminator"}},
							BinaryName:     "ziti-router",
							ConfigSrc:      "ingress_router.yml",
							ConfigName:     "001.yml",
							PublicIdentity: "001",
						},
					},
				},
			},
		},
	},
}
