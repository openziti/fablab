package model

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

type testTemplateStrategy struct {
}

func (t testTemplateStrategy) IsTemplated(entity Entity) bool {
	return entity.GetScope().HasTag("templated")
}

func (t testTemplateStrategy) GetEntityCount(entity Entity) int {
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
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"single-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"scaled-component-{{ .Index }}": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
						},
					},
					"single-host": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "templated"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"single-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"scaled-component-{{ .Index }}": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
						},
					},
					"scaled-host-{{ .Index }}": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "templated"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"single-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"scaled-component-{{ .Index }}": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
						},
					},
					"dropped-host": {
						Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
					},
				},
			},
			"single-region": {
				Scope:  Scope{Tags: Tags{"a", "b.{{ .Index }}", "templated"}},
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
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"single-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"scaled-component-{{ .Index }}": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
						},
					},
					"single-host": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "templated"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"single-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"scaled-component-{{ .Index }}": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
						},
					},
					"scaled-host-{{ .Index }}": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "templated"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"single-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"scaled-component-{{ .Index }}": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
						},
					},
					"dropped-host": {
						Scope: Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
					},
				},
			},
			"scaled-region-{{ .Index }}": {
				Scope:  Scope{Tags: Tags{"a", "b.{{ .Index }}", "templated"}},
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
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"single-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"scaled-component-{{ .Index }}": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
						},
					},
					"single-host": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "templated"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"single-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"scaled-component-{{ .Index }}": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
						},
					},
					"scaled-host-{{ .Index }}": {
						Scope:                Scope{Tags: Tags{"c", "d.{{ .Index}}.{{ .Region.Id }}", "templated"}},
						PublicIp:             "1.1.{{ .Index}}.{{ .Region.Id }}",
						PrivateIp:            "2.2.{{ .Index}}.{{ .Region.Id }}",
						InstanceType:         "type-{{ .Index}}.{{ .Region.Id }}", // should not get interpreted
						InstanceResourceType: "resource-{{ .Index}}.{{ .Region.Id }}",
						SpotPrice:            "spot-price-{{ .Index}}.{{ .Region.Id }}",
						SpotType:             "spot-type-{{ .Index}}.{{ .Region.Id }}",
						Components: Components{
							"static-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"single-component": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
							"scaled-component-{{ .Index }}": {
								Scope:           Scope{Tags: Tags{"e", "f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", "templated"}},
								ScriptSrc:       "script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ScriptName:      "script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigSrc:       "config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								ConfigName:      "config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								BinaryName:      "binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PublicIdentity:  "public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
								PrivateIdentity: "private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}",
							},
						},
					},
					"dropped-host": {
						Scope: Scope{Tags: Tags{"templated"}},
					},
				},
			},
		},
	}

	model.init("test")

	factory := &TemplatingFactory{
		Strategy: testTemplateStrategy{},
	}

	req := require.New(t)
	req.Equal(9, len(model.SelectHosts(".templated")))
	req.Equal(18, len(model.SelectComponents(".templated")))

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
	if region.HasTag("templated") {
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
	if host.HasTag("templated") {
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
	if component.HasTag("templated") {
		req.Equal(fmt.Sprintf("f-%v.%v.%v", component.Index, host.Id, region.Id), component.Tags[1])
		req.Equal(fmt.Sprintf("script-src-%v.%v.%v", component.Index, host.Id, region.Id), component.ScriptSrc)
		req.Equal(fmt.Sprintf("script-name-%v.%v.%v", component.Index, host.Id, region.Id), component.ScriptName)
		req.Equal(fmt.Sprintf("config-src-%v.%v.%v", component.Index, host.Id, region.Id), component.ConfigSrc)
		req.Equal(fmt.Sprintf("config-name-%v.%v.%v", component.Index, host.Id, region.Id), component.ConfigName)
		req.Equal(fmt.Sprintf("binary-%v.%v.%v", component.Index, host.Id, region.Id), component.BinaryName)
		req.Equal(fmt.Sprintf("public-id-%v.%v.%v", component.Index, host.Id, region.Id), component.PublicIdentity)
		req.Equal(fmt.Sprintf("private-id-%v.%v.%v", component.Index, host.Id, region.Id), component.PrivateIdentity)
	} else {
		req.Equal("f-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", component.Tags[1])
		req.Equal("script-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", component.ScriptSrc)
		req.Equal("script-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", component.ScriptName)
		req.Equal("config-src-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", component.ConfigSrc)
		req.Equal("config-name-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", component.ConfigName)
		req.Equal("binary-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", component.BinaryName)
		req.Equal("public-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", component.PublicIdentity)
		req.Equal("private-id-{{ .Index }}.{{ .Host.Id }}.{{ .Region.Id }}", component.PrivateIdentity)
	}
}
