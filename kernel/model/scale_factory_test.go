package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type testScaleStrategy struct {
}

func (t testScaleStrategy) IsScaled(entity Entity) bool {
	return entity.GetScope().HasTag("scaled")
}

func (t testScaleStrategy) GetEntityCount(entity Entity) uint32 {
	if strings.HasPrefix(entity.GetId(), "dropped") {
		return 0
	}
	if strings.HasPrefix(entity.GetId(), "single") {
		return 1
	}
	return 3
}

func Test_Templating(t *testing.T) {
	model := &Model{
		Id: "test",
		Regions: Regions{
			"static-region": {
				Scope:  Scope{Tags: Tags{"a", "b.{{ .Index }}"}},
				Region: "us-west-1",
				Site:   "us-west-1-{{ .Index }}",
				Hosts: Hosts{
					"static-host": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}",
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
							},
							"single-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
							"scaled-component-{{ .ScaleIndex }}": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
						},
					},
					"single-host": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "scaled"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
							},
							"single-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
							"scaled-component-{{ .ScaleIndex }}": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
						},
					},
					"scaled-host-{{ .ScaleIndex }}": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "scaled"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
							},
							"single-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
							"scaled-component-{{ .ScaleIndex }}": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
						},
					},
					"dropped-host": {
						Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
					},
				},
			},
			"single-region": {
				Scope:  Scope{Tags: Tags{"a", "b.{{ .Index }}", "scaled"}},
				Region: "us-west-1",
				Site:   "us-west-1-{{ .Index }}", // should get templated
				Hosts: Hosts{
					"static-host": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}",
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
							},
							"single-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
							"scaled-component-{{ .ScaleIndex }}": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
						},
					},
					"single-host": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "scaled"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
							},
							"single-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
							"scaled-component-{{ .ScaleIndex }}": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
						},
					},
					"scaled-host-{{ .ScaleIndex }}": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "scaled"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
							},
							"single-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
							"scaled-component-{{ .ScaleIndex }}": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
						},
					},
					"dropped-host": {
						Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
					},
				},
			},
			"scaled-region-{{ .ScaleIndex }}": {
				Scope:  Scope{Tags: Tags{"a", "b.{{ .Index }}", "scaled"}},
				Region: "us-west-1",
				Site:   "us-west-1-{{ .Index }}",
				Hosts: Hosts{
					"static-host": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}",
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
							},
							"single-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
							"scaled-component-{{ .ScaleIndex }}": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
						},
					},
					"single-host": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "scaled"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
							},
							"single-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
							"scaled-component-{{ .ScaleIndex }}": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
						},
					},
					"scaled-host-{{ .ScaleIndex }}": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "scaled"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
							},
							"single-component": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
							"scaled-component-{{ .ScaleIndex }}": {
								Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "scaled"}},
							},
						},
					},
					"dropped-host": {
						Scope: Scope{Tags: Tags{"scaled"}},
					},
				},
			},
		},
	}

	req := require.New(t)
	req.NoError(model.init())

	factory := &ScaleFactory{
		Strategy:      testScaleStrategy{},
		EntityFactory: DefaultScaleEntityFactory{},
	}

	req.Equal(9, len(model.SelectHosts(".scaled")))
	req.Equal(18, len(model.SelectComponents(".scaled")))

	req.NoError(factory.Build(model))

	req.Equal(5, len(model.Regions))

	validateRegion(req, model, model.Regions["static-region"])
	validateRegion(req, model, model.Regions["single-region"])
	validateRegion(req, model, model.Regions["scaled-region-0"])
	validateRegion(req, model, model.Regions["scaled-region-1"])
	validateRegion(req, model, model.Regions["scaled-region-2"])
}

func validateRegion(req *require.Assertions, model *Model, region *Region) {
	req.NotNil(region)
	req.Equal(model, region.Model)

	req.Equal("us-west-1", region.Region)
	req.True(len(region.Tags) >= 2)
	req.Equal("a", region.Tags[0])
	if region.HasTag("scaled") {
		req.Equal(fmt.Sprintf("us-west-1-%v", region.Index), region.Site)
		req.Equal(fmt.Sprintf("b.%v", region.Index), region.Tags[1])
	} else {
		req.Equal("us-west-1-{{ .Index }}", region.Site)
		req.Equal("b.{{ .Index }}", region.Tags[1])
	}

	req.Equal(5, len(region.Hosts))
	validateHost(req, region, region.Hosts["static-host"])
	validateHost(req, region, region.Hosts["single-host"])
	validateHost(req, region, region.Hosts["scaled-host-0"])
	validateHost(req, region, region.Hosts["scaled-host-1"])
	validateHost(req, region, region.Hosts["scaled-host-2"])
}

func validateHost(req *require.Assertions, region *Region, host *Host) {
	req.NotNil(host)
	req.Equal(region, host.Region)

	req.True(len(host.Tags) >= 2)
	req.Equal("c", host.Tags[0])
	if host.HasTag("scaled") || region.HasTag("scaled") {
		req.Equal(fmt.Sprintf("d.%v.%v", host.Index, region.Id), host.Tags[1])
		req.Equal(fmt.Sprintf("1.1.%v.%v", host.Index, region.Id), host.PublicIp)
		req.Equal(fmt.Sprintf("2.2.%v.%v", host.Index, region.Id), host.PrivateIp)
		req.Equal(fmt.Sprintf("type-%v.%v", host.Index, region.Id), host.InstanceType)
		req.Equal(fmt.Sprintf("resource-%v.%v", host.Index, region.Id), host.InstanceResourceType)
		req.Equal(fmt.Sprintf("spot-price-%v.%v", host.Index, region.Id), host.SpotPrice)
		req.Equal(fmt.Sprintf("spot-type-%v.%v", host.Index, region.Id), host.SpotType)
	} else {
		req.Equal("d.{{ .Index}}.{{ .Region.Id }}", host.Tags[1])
		req.Equal("1.1.{{ .Index}}.{{ .Region.Id }}", host.PublicIp)
		req.Equal("2.2.{{ .Index}}.{{ .Region.Id }}", host.PrivateIp)
		req.Equal("type-{{ .Index}}.{{ .Region.Id }}", host.InstanceType)
		req.Equal("resource-{{ .Index}}.{{ .Region.Id }}", host.InstanceResourceType)
		req.Equal("spot-price-{{ .Index}}.{{ .Region.Id }}", host.SpotPrice)
		req.Equal("spot-type-{{ .Index}}.{{ .Region.Id }}", host.SpotType)
	}

	req.Equal(5, len(region.Hosts))
	validateComponent(req, region, host, host.Components["static-component"])
	validateComponent(req, region, host, host.Components["single-component"])
	validateComponent(req, region, host, host.Components["scaled-component-0"])
	validateComponent(req, region, host, host.Components["scaled-component-1"])
	validateComponent(req, region, host, host.Components["scaled-component-2"])
}

func validateComponent(req *require.Assertions, region *Region, host *Host, component *Component) {
	req.NotNil(component)
	req.Equal(host, component.Host)

	req.True(len(component.Tags) >= 2)
	req.Equal("e", component.Tags[0])
	if component.HasTag("scaled") || host.HasTag("scaled") || region.HasTag("scaled") {
		req.Equal(fmt.Sprintf("f-%v.%v.%v", component.Index, host.Id, region.Id), component.Tags[1])
	} else {
		req.Equal("f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", component.Tags[1])
	}
}
