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

package zitilib_characterization

import (
	"github.com/openziti/fablab/kernel/fablib/binding"
	"github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/aws_ssh_key"
	"github.com/openziti/fablab/kernel/model"
)

func init() {
	model.RegisterModel("zitilib/characterization", Ziti)
	model.AddBootstrapExtension(binding.AwsCredentialsLoader)
	model.AddBootstrapExtension(aws_ssh_key.KeyManager)
}

// Static model skeleton for zitilib/characterization
//
var Ziti = &model.Model{
	Scope: model.Scope{
		Variables: model.Variables{
			"zitilib": model.Variables{
				"fabric": model.Variables{
					"data_plane_protocol": &model.Variable{Default: "tls"},
				},
			},
			"characterization": model.Variables{
				"sample_minutes": &model.Variable{Default: 1},
				"tcpdump": model.Variables{
					"enabled": &model.Variable{Default: true},
					"snaplen": &model.Variable{Default: 128},
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
					"username": &model.Variable{Default: "fedora"},
				},
			},
			"distribution": model.Variables{
				"rsync_bin": &model.Variable{Default: "rsync"},
				"ssh_bin":   &model.Variable{Default: "ssh"},
			},
		},
	},

	Factories: []model.Factory{
		newHostsFactory(),
		newActionsFactory(),
		newStagesFactory(),
	},

	Regions: model.Regions{
		"local": {
			Region: "us-east-1",
			Site:   "us-east-1a",
			Hosts: model.Hosts{
				"ctrl": {
					Scope: model.Scope{Tags: model.Tags{"^ctrl"}},
					Components: model.Components{
						"ctrl": {
							BinaryName:     "ziti-controller",
							ConfigSrc:      "ctrl.yml",
							ConfigName:     "ctrl.yml",
							PublicIdentity: "ctrl",
						},
					},
				},
				"local": {
					Scope: model.Scope{Tags: model.Tags{"^router", "^terminator"}},
					Components: model.Components{
						"local": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "egress_router.yml",
							ConfigName:     "local.yml",
							PublicIdentity: "local",
						},
					},
				},
				"service": {
					Scope: model.Scope{Tags: model.Tags{"service", "iperf_server"}},
				},
			},
		},
		"short": {
			Region: "us-west-1",
			Site:   "us-west-1c",
			Hosts: model.Hosts{
				"short": {
					Scope: model.Scope{Tags: model.Tags{"^router"}},
					Components: model.Components{
						"short": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "ingress_router.yml",
							ConfigName:     "short.yml",
							PublicIdentity: "short",
						},
					},
				},
				"client": {
					Scope: model.Scope{Tags: model.Tags{"client", "iperf_client"}},
				},
			},
		},
		"medium": {
			Region: "ap-south-1",
			Site:   "ap-south-1a",
			Hosts: model.Hosts{
				"medium": {
					Scope: model.Scope{Tags: model.Tags{"^router"}},
					Components: model.Components{
						"medium": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "ingress_router.yml",
							ConfigName:     "medium.yml",
							PublicIdentity: "medium",
						},
					},
				},
				"client": {
					Scope: model.Scope{Tags: model.Tags{"client", "iperf_client"}},
				},
			},
		},
		"long": {
			Region: "ap-southeast-2",
			Site:   "ap-southeast-2c",
			Hosts: model.Hosts{
				"long": {
					Scope: model.Scope{Tags: model.Tags{"^router"}},
					Components: model.Components{
						"long": {
							BinaryName:     "ziti-router",
							ConfigSrc:      "ingress_router.yml",
							ConfigName:     "long.yml",
							PublicIdentity: "long",
						},
					},
				},
				"client": {
					Scope: model.Scope{Tags: model.Tags{"client", "iperf_client"}},
				},
			},
		},
	},
}
