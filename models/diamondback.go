package models

import (
	"github.com/netfoundry/fablab/kernel"
)

var diamondback = &kernel.Model{
	Scope: kernelScope,
	Regions: kernel.Regions{
		"initiator": {
			Scope: kernel.Scope{
				Tags: kernel.Tags{"ctrl", "router", "loop", "initiator"},
			},
			Id: "us-east-1",
			Az: "us-east-1a",
			Hosts: kernel.Hosts{
				"ctrl": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"ctrl"},
						Variables: kernel.Variables{"instance_type": instanceType("m5.large")},
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
					},
				},
				"001": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"router", "initiator"},
						Variables: kernel.Variables{"instance_type": instanceType("m5.large")},
					},
					Components: kernel.Components{
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
				"loop0": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"loop-dialer"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.medium")},
					},
				},
				"loop1": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"loop-dialer"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.medium")},
					},
				},
				"loop2": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"loop-dialer"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.medium")},
					},
				},
				"loop3": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"loop-dialer"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.medium")},
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
						Variables: kernel.Variables{"instance_type": instanceType("m5.large")},
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
						Variables: kernel.Variables{"instance_type": instanceType("m5.large")},
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
				Tags: kernel.Tags{"router", "loop", "terminator"},
			},
			Id: "us-west-2",
			Az: "us-west-2b",
			Hosts: kernel.Hosts{
				"003": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"router"},
						Variables: kernel.Variables{"instance_type": instanceType("m5.large")},
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
				"loop0": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"loop-listener"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.micro")},
					},
				},
				"loop1": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"loop-listener"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.micro")},
					},
				},
				"loop2": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"loop-listener"},
						Variables: kernel.Variables{"instance_type": instanceType("t2.micro")},
					},
				},
				"loop3": {
					Scope: kernel.Scope{
						Tags:      kernel.Tags{"loop-listener"},
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
	Disposal:       commonDisposal(),
}
