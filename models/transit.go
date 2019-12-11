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

var transit = &kernel.Model{
	Scope: kernelScope,
	Regions: kernel.Regions{
		"initiator": {
			Scope: kernel.Scope{
				Tags: kernel.Tags{"initiator", "ctrl", "router", "iperf-client"},
			},
			Id: "us-east-1",
			Az: "us-east-1a",
			Hosts: kernel.Hosts{
				"001": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"ctrl", "router", "initiator"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.micro")},
					},
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
								Tags: kernel.Tags{"router"},
							},
							BinaryName:     "ziti-router",
							ConfigSrc:      "ingress_router.yml",
							ConfigName:     "001.yml",
							PublicIdentity: "001",
						},
					},
				},
				"iperf-client": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"iperf-client"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.micro")},
					},
				},
			},
		},
		"transitA": {
			Scope: kernel.Scope{
				Tags: kernel.Tags{"router"},
			},
			Id: "us-west-1",
			Az: "us-west-1b",
			Hosts: kernel.Hosts{
				"002": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"router"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.micro")},
					},
					Components: kernel.Components{
						"002": {
							Scope: kernel.Scope{
								Tags: kernel.Tags{"router"},
							},
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
			Scope: kernel.Scope{
				Tags: kernel.Tags{"router"},
			},
			Id: "us-east-2",
			Az: "us-east-2c",
			Hosts: kernel.Hosts{
				"004": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"router"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.micro")},
					},
					Components: kernel.Components{
						"004": {
							Scope: kernel.Scope{
								Tags: kernel.Tags{"router"},
							},
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
			Scope: kernel.Scope{
				Tags: kernel.Tags{"router", "terminator", "iperf-server"},
			},
			Id: "us-west-2",
			Az: "us-west-2b",
			Hosts: kernel.Hosts{
				"003": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"router"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.micro")},
					},
					Components: kernel.Components{
						"003": {
							Scope: kernel.Scope{
								Tags: kernel.Tags{"router", "terminator"},
							},
							BinaryName:     "ziti-router",
							ConfigSrc:      "egress_router.yml",
							ConfigName:     "003.yml",
							PublicIdentity: "003",
						},
					},
				},
				"iperf-server": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"iperf-server"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.micro")},
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
	Operation:      commonOperation(),
	Disposal:       commonDisposal(),
}
