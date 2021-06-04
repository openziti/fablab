// +build vartest

package test

import (
	"fmt"
	"github.com/openziti/fablab/kernel/model"
)

func init() {
	model.RegisterModel("test/vartest", vartest)
}

var vartest = &model.Model{
	Factories: []model.Factory{
		stagesFactory{},
	},
	Scope: model.Scope{
		Defaults: model.Variables{
			"test": model.Variables{
				"key": "model.hello",
			},
		},
	},
	Regions: model.Regions{
		"region1": {
			Scope: model.Scope{
				Defaults: model.Variables{
					"test": model.Variables{
						"key": "region.hello",
					},
				},
			},
			Hosts: model.Hosts{
				"host1": {
					Scope: model.Scope{
						Defaults: model.Variables{
							"test": model.Variables{
								"key": "host.hello",
							},
						},
					},
					Components: model.Components{
						"component1": {
							Scope: model.Scope{
								Defaults: model.Variables{
									"test": model.Variables{
										"key": "hello",
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

type stagesFactory struct{}

func (s stagesFactory) Build(m *model.Model) error {
	m.AddOperatingStageF(func(run model.Run) error {
		region := m.Regions["region1"]
		host := region.Hosts["host1"]
		component := host.Components["component1"]

		fmt.Printf("Model value    : %v\n", m.MustStringVariable("test.key"))
		fmt.Printf("Region value   : %v\n", region.MustStringVariable("test.key"))
		fmt.Printf("Host value     : %v\n", host.MustStringVariable("test.key"))
		fmt.Printf("Component value: %v\n", component.MustStringVariable("test.key"))
		return nil
	})
	return nil
}
