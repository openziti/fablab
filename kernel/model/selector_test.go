package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestModel() *Model {
	return &Model{
		Id:    "test",
		Scope: Scope{Tags: Tags{"global"}},
		Regions: Regions{
			"initiator": {
				Scope:  Scope{Tags: Tags{"region-shared", "region-first", "region-not-inherited"}},
				Region: "us-east-1",
				Site:   "us-east-1a",
				Hosts: Hosts{
					"ctrl": {
						Scope: Scope{Tags: Tags{"ctrl"}},
						Components: Components{
							"ctrl": {
								Scope: Scope{Tags: Tags{"ctrl"}},
							},
						},
					},
					"initiator": {
						Scope: Scope{Tags: Tags{"initiator", "edge-router"}},
						Components: Components{
							"initiator": {
								Scope: Scope{Tags: Tags{"edge-router", "initiator"}},
							},
						},
					},
					"client": {
						Scope: Scope{Tags: Tags{"client", "sdk-app"}},
						Components: Components{
							"client1": {},
						},
					},
				},
			},
			"terminator": {
				Region: "us-west-1",
				Site:   "us-west-1b",
				Scope:  Scope{Tags: Tags{"region-shared", "region-last"}},
				Hosts: Hosts{
					"terminator": {
						Scope: Scope{Tags: Tags{"terminator", "edge-router"}},
						Components: Components{
							"terminator": {
								Scope: Scope{Tags: Tags{"edge-router", "terminator"}},
							},
						},
					},
					"service": {
						Scope: Scope{Tags: Tags{"service", "sdk-app"}},
						Components: Components{
							"server1": {},
						},
					},
				},
			},
		},
	}
}

func TestModel_SelectRegions(t *testing.T) {
	req := require.New(t)
	model := createTestModel()
	req.NoError(model.init())

	// test lookup without prefix. Should be like lookup by id
	regions := model.SelectRegions("initiator")
	req.Equal(1, len(regions))
	req.Equal("initiator", regions[0].GetId())
	req.Equal("us-east-1", regions[0].Region)

	// test lookup by id
	regions = model.SelectRegions("#initiator")
	req.Equal(1, len(regions))
	req.Equal("initiator", regions[0].GetId())
	req.Equal("us-east-1", regions[0].Region)

	regions = model.SelectRegions(".region-not-inherited")
	req.Equal(1, len(regions))
	req.Equal("initiator", regions[0].GetId())
	req.Equal("us-east-1", regions[0].Region)

	regions = model.SelectRegions("*")
	req.Equal(2, len(regions))
	req.NotEqual(regions[0].GetId(), regions[1].GetId())

	regions = model.SelectRegions("^.global")
	req.Equal(2, len(regions))
	req.NotEqual(regions[0].GetId(), regions[1].GetId())

	regions = model.SelectRegions(".region-shared")
	req.Equal(2, len(regions))
	req.NotEqual(regions[0].GetId(), regions[1].GetId())

	regions = model.SelectRegions(".region-first")
	req.Equal(1, len(regions))

	regions = model.SelectRegions(".region-first, .region-last")
	req.Equal(2, len(regions))
	req.NotEqual(regions[0].GetId(), regions[1].GetId())

	regions = model.SelectRegions("component.ctrl")
	req.Equal(1, len(regions))
	req.Equal("initiator", regions[0].GetId())

	regions = model.SelectRegions("component.ctrl")
	req.Equal(1, len(regions))
	req.Equal("initiator", regions[0].GetId())

	regions = model.SelectRegions("host.service")
	req.Equal(1, len(regions))
	req.Equal("terminator", regions[0].GetId())

	regions = model.SelectRegions("host#service")
	req.Equal(1, len(regions))
	req.Equal("terminator", regions[0].GetId())

	regions = model.SelectRegions("host.sdk-app")
	req.Equal(2, len(regions))
	req.NotEqual(regions[0].GetId(), regions[1].GetId())

	regions = model.SelectRegions("component#ctrl")
	req.Equal(1, len(regions))
	req.Equal("initiator", regions[0].GetId())

	regions = model.SelectRegions("component#ctrl.ctrl")
	req.Equal(1, len(regions))
	req.Equal("initiator", regions[0].GetId())

	regions = model.SelectRegions("component#ctrl.terminator")
	req.Equal(0, len(regions))
}

func TestModel_SelectHosts(t *testing.T) {
	req := require.New(t)
	model := createTestModel()
	req.NoError(model.init())

	// test lookup by id
	hosts := model.SelectHosts("#ctrl")
	req.Equal(1, len(hosts))
	req.Equal("ctrl", hosts[0].GetId())

	hosts = model.SelectHosts("*")
	req.Equal(5, len(hosts))

	// ensure tags are inherited
	hosts = model.SelectHosts("^.global")
	req.Equal(5, len(hosts))

	hosts = model.SelectHosts("^.region-shared")
	req.Equal(5, len(hosts))

	hosts = model.SelectHosts("^.region-first")
	req.Equal(3, len(hosts))

	hosts = model.SelectHosts("model.global > *")
	req.Equal(5, len(hosts))

	hosts = model.SelectHosts("#initiator > *")
	req.Equal(3, len(hosts))

	hosts = model.SelectHosts(".region-first > *")
	req.Equal(3, len(hosts))

	hosts = model.SelectHosts(".edge-router")
	req.Equal(2, len(hosts))

	hosts = model.SelectHosts("#initiator > .edge-router")
	req.Equal(1, len(hosts))
	req.Equal("initiator", hosts[0].GetId())

	hosts = model.SelectHosts("#initiator > #ctrl")
	req.Equal(1, len(hosts))
	req.Equal("ctrl", hosts[0].GetId())

	hosts = model.SelectHosts("#initiator > #terminator")
	req.Equal(0, len(hosts))

	hosts = model.SelectHosts(".region-not-inherited > *")
	req.Equal(3, len(hosts))

	hosts = model.SelectHosts(".region-not-inherited")
	req.Equal(0, len(hosts))

	hosts = model.SelectHosts("region.region-first")
	req.Equal(3, len(hosts))

	hosts = model.SelectHosts("region#initiator")
	req.Equal(3, len(hosts))

	hosts = model.SelectHosts("region#initiator.region-first")
	req.Equal(3, len(hosts))

	hosts = model.SelectHosts("region.region-first, region.region-last")
	req.Equal(5, len(hosts))

	hosts = model.SelectHosts("region.region-first .ctrl component#ctrl")
	req.Equal(1, len(hosts))

	hosts = model.SelectHosts("region.region-first .ctrl component#foo")
	req.Equal(0, len(hosts))
}

func TestModel_SelectComponents(t *testing.T) {
	req := require.New(t)
	model := createTestModel()
	req.NoError(model.init())

	// test lookup by id
	components := model.SelectComponents("#terminator")
	req.Equal(1, len(components))
	req.Equal("terminator", components[0].GetId())

	components = model.SelectComponents("*")
	req.Equal(5, len(components))

	components = model.SelectComponents("^.global")
	req.Equal(5, len(components))

	components = model.SelectComponents("model.global > *")
	req.Equal(5, len(components))

	components = model.SelectComponents("model.global > * > *")
	req.Equal(5, len(components))

	components = model.SelectComponents(".region-shared > * > *")
	req.Equal(5, len(components))

	components = model.SelectComponents(".region-first > * > *")
	req.Equal(3, len(components))

	components = model.SelectComponents("^.region-first")
	req.Equal(3, len(components))

	components = model.SelectComponents("* > region.region-first > *")
	req.Equal(3, len(components))

	components = model.SelectComponents("region.region-first > *")
	req.Equal(3, len(components))

	components = model.SelectComponents("^.region-first")
	req.Equal(3, len(components))

	components = model.SelectComponents("region.region-first > host.sdk-app")
	req.Equal(1, len(components))
	req.Equal("client1", components[0].GetId())

	components = model.SelectComponents("#initiator > #client > host.sdk-app")
	req.Equal(1, len(components))
	req.Equal("client1", components[0].GetId())

	components = model.SelectComponents("#initiator > #client > #client1")
	req.Equal(1, len(components))
	req.Equal("client1", components[0].GetId())

	components = model.SelectComponents(".global > .region-first > * > host.client.sdk-app")
	req.Equal(1, len(components))
	req.Equal("client1", components[0].GetId())

	components = model.SelectComponents("region#initiator host.ctrl")
	req.Equal(1, len(components))

	components = model.SelectComponents("region#initiator")
	req.Equal(3, len(components))

	components = model.SelectComponents("region#initiator component.ctrl")
	req.Equal(1, len(components))
}
