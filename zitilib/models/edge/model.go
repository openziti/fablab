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

package edge

import (
	"github.com/openziti/fablab/kernel/fablib/binding"
	"github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/aws_ssh_key"
	"github.com/openziti/fablab/kernel/model"
	"strings"
)

func init() {
	model.RegisterModel("zitilib/edge", edge)
	model.AddBootstrapExtension(binding.AwsCredentialsLoader)
	model.AddBootstrapExtension(aws_ssh_key.KeyManager)
}

type templateStrategy struct{}

func (t templateStrategy) IsTemplated(entity model.Entity) bool {
	return strings.Contains(entity.GetId(), ".Index")
}

func (t templateStrategy) GetEntityCount(entity model.Entity) int {
	if entity.GetType() == model.EntityTypeHost {
		if entity.GetScope().HasTag("service") {
			return 3
		}
		return 3
	}
	return 1
}

// Static model skeleton for zitilib/edge
//
var edge = &model.Model{
	//Scope: model.BuildScope().
	//	Var("edge", "region").Default("us-east-1").
	//	Var("edge", "az").Default("us-east-1c").
	//	Var("edge", "sizing", "ctrl").Default("t2-medium").
	//	Var("edge", "sizing", "initiator").Default("t2-medium").
	//	Var("edge", "sizing", "terminator").Default("t2-medium").
	//	Var("edge", "sizing", "client").Default("t2-medium").
	//	Var("edge", "sizing", "service").Default("t2-medium").
	//	Var("zitilib", "fabric", "data_plane_protocol").Default("tls").
	//	Var("environment").Required().
	//	Var("credentials", "aws", "access_key").Required().Sensitive().
	//	Var("credentials", "aws", "secret_key").Required().Sensitive().
	//	Var("credentials", "aws", "ssh_key_name").Required().
	//	Var("credentials", "ssh", "key_path").Required().
	//	Var("credentials", "ssh", "username").Default("fedora").
	//	Var("credentials", "edge", "username").Required().Sensitive().
	//	Var("credentials", "edge", "password").Required().Sensitive().
	//	Var("credentials", "influxdb", "username").Required().Sensitive().
	//	Var("credentials", "influxdb", "password").Required().Sensitive().
	//	Build(),
	Scope: model.Scope{
		Variables: model.Variables{
			"edge": model.Variables{
				"region": &model.Variable{Default: "us-east-1"},
				"az":     &model.Variable{Default: "us-east-1c"},
				"sizing": model.Variables{
					"ctrl":       &model.Variable{Default: "t3a.medium"},
					"initiator":  &model.Variable{Default: "t3a.medium"},
					"terminator": &model.Variable{Default: "t3a.medium"},
					"client":     &model.Variable{Default: "t3a.medium"},
					"service":    &model.Variable{Default: "t3a.medium"},
				},
			},
			"zitilib": model.Variables{ //configs in /lib link to this....
				"fabric": model.Variables{
					"data_plane_protocol": &model.Variable{Default: "tls"},
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
					"username": &model.Variable{Default: "ubuntu"},
				},
				"edge": model.Variables{
					"username": &model.Variable{Required: true, Sensitive: true},
					"password": &model.Variable{Required: true, Sensitive: true},
				},
				"influxdb": model.Variables{
					"username": &model.Variable{Required: true, Sensitive: true},
					"password": &model.Variable{Required: true, Sensitive: true},
				},
			},
			"distribution": model.Variables{
				"rsync_bin": &model.Variable{Default: "rsync"},
				"ssh_bin":   &model.Variable{Default: "ssh"},
			},
			"metrics": model.Variables{
				"influxdb": model.Variables{
					"url": &model.Variable{Default: "http://localhost:8086"},
					"db":  &model.Variable{Default: "ziti"},
				},
			},
		},
	},

	ModelFactories: []model.Factory{
		&model.TemplatingFactory{Strategy: templateStrategy{}},
	},

	Factories: []model.Factory{
		newHostsFactory(),
		newActionsFactory(),
		newStageFactory(),
	},

	Regions: model.Regions{
		"initiator": {
			Region: "us-east-1",
			Site:   "us-east-1c",
			Hosts: model.Hosts{
				"ctrl": {
					Scope: model.Scope{Tags: model.Tags{"^ctrl"}},
					Components: model.Components{
						"ctrl": {
							BinaryName:     "ziti-controller",
							ConfigSrc:      "ctrl_edge.yml",
							ConfigName:     "ctrl_edge.yml",
							PublicIdentity: "ctrl",
						},
					},
				},
				"initiator": {
					Scope: model.Scope{Tags: model.Tags{"^edge-router", "^initiator", "^perf-test"}},
					Components: model.Components{
						"initiator": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "edge_router.yml",
							ConfigName:     "edge_router_initiator.yml",
							PublicIdentity: "edge_router_initiator",
						},
					},
				},
				"metricsRouter": {
					Scope: model.Scope{Tags: model.Tags{"^edge-router", "^metrics"}},
					Components: model.Components{
						"initiator": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "edge_router_isolated.yml",
							ConfigName:     "edge_router_metrics.yml",
							PublicIdentity: "edge_router_metrics",
						},
					},
				},
				"client{{ .Index }}": {
					Scope: model.Scope{Tags: model.Tags{"^client", "^sdk-app"}},
					Components: model.Components{
						"client{{ .Host.Index }}": {
							BinaryName:     "ziti-fabric-test",
							PublicIdentity: "client{{ .Host.Index }}",
							ConfigSrc:      "loop/edge-perf.loop3.yml",
							ConfigName:     "edge-perf-{{ .Host.Index }}.loop3.yml",
						},
					},
				},
			},
		},
		"terminator": {
			Region: "us-west-1",
			Site:   "us-west-1c",
			Hosts: model.Hosts{
				"terminator": {
					Scope: model.Scope{Tags: model.Tags{"^edge-router", "^terminator", "^perf-test"}},
					Components: model.Components{
						"terminator": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "edge_router.yml",
							ConfigName:     "edge_router_terminator.yml",
							PublicIdentity: "edge_router_terminator",
						},
					},
				},
				"service{{ .Index }}": {
					Scope: model.Scope{Tags: model.Tags{"^service", "^sdk-app"}},
					Components: model.Components{
						"server{{ .Host.Index }}": {
							BinaryName:     "ziti-fabric-test",
							PublicIdentity: "server{{ .Host.Index }}",
						},
					},
				},
			},
		},
	},
}
