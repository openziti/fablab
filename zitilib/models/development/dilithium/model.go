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

package dilithium

import (
	"github.com/openziti/fablab/kernel/fablib/binding"
	"github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/aws_ssh_key"
	"github.com/openziti/fablab/kernel/model"
)

func init() {
	model.RegisterModel("zitilib/development/dilithium", dilithiumModel)
	model.AddBootstrapExtension(binding.AwsCredentialsLoader)
	model.AddBootstrapExtension(aws_ssh_key.KeyManager)
}

var dilithiumModel = &model.Model{
	Regions: model.Regions{
		"local": {
			Scope: model.Scope{Tags: model.Tags{"virginia"}},
			Hosts: model.Hosts{
				"host": {InstanceType: "t2.medium"},
			},
			Region: "us-east-1",
			Site:   "us-east-1a",
		},
		"remote": {
			Scope: model.Scope{Tags: model.Tags{"california"}},
			Hosts: model.Hosts{
				"host": {InstanceType: "t2.medium"},
			},
			Region: "us-west-1",
			Site:   "us-west-1c",
		},
	},

	Scope: model.Scope{
		Variables: model.Variables{
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
			"instance_type": &model.Variable{Default: "t2.medium"},
		},
	},

	Factories: []model.Factory{
		newHostsFactory(),
		newActionsFactory(),
		newStagesFactory(),
	},
	BootstrapExtensions: []model.BootstrapExtension{
		&bootstrap{},
	},
}
