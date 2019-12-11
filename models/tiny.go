/*
	Copyright 2019 Netfoundry, Inc.

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

package models

import (
	"github.com/netfoundry/fablab/kernel"
)

var tiny = &kernel.Model{
	Scope: kernelScope,

	Regions: kernel.Regions{
		"tiny": {
			Scope: kernel.Scope{
				Tags: kernel.Tags{"ctrl", "router", "loop", "initiator", "terminator"},
			},
			Id: "us-east-1",
			Az: "us-east-1c",
			Hosts: kernel.Hosts{
				"loop0": {
					Scope: kernel.Scope{
						Tags: kernel.Tags{"ctrl", "router", "loop-dialer", "loop-listener", "initiator", "terminator"},
					},
					InstanceType: "m5.large",
					Components: kernel.Components{
						"ctrl": {
							Scope: kernel.Scope{
								Tags: kernel.Tags{"ctrl"},
							},
							BinaryName:     "ziti-controller",
							ConfigSrc:      "ctrl.yml",
							ConfigName:     "ctrl.yml",
							PublicIdentity: "ctrl",
						},
						"001": {
							Scope: kernel.Scope{
								Tags: kernel.Tags{"router", "terminator"},
							},
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

	Actions:        commonActions(),
	Infrastructure: commonInfrastructure(),
	Configuration:  commonConfiguration(),
	Kitting:        commonKitting(),
	Distribution:   commonDistribution(),
	Activation:     commonActivation(),
	Disposal:       commonDisposal(),
}
