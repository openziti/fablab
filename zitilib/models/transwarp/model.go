/*
	Copyright NetFoundry, Inc.

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

package transwarp

import "github.com/openziti/fablab/kernel/model"

func init() {
	model.RegisterModel("zitilib/transwarp", transwarpModel)
}

var transwarpModel = &model.Model{
	Regions: model.Regions{
		"local": {
			Hosts: model.Hosts{
				"ctrl":    {},
				"router":  {},
				"service": {},
			},
			Id: "us-east-1",
			Az: "us-east-1c",
		},
		"remote": {
			Hosts: model.Hosts{
				"router": {},
				"client": {},
			},
			Id: "us-west-1",
			Az: "us-west-1a",
		},
	},

	Scope: model.Scope{
		Variables: model.Variables{
			"zitilib": model.Variables{
				"fabric": model.Variables{
					"data_plane_protocol": &model.Variable{Default: "transwarp"},
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
		newInfrastructureFactory(),
	},
}
