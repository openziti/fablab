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
