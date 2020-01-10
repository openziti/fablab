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

package zitilab_characterization_ziti

import "github.com/netfoundry/fablab/kernel/model"

func init() {
	model.RegisterModel("zitilab/characterization/ziti", Model)
}

// Static model skeleton for zitilab/characterization/ziti
//
var Model = &model.Model{
	Factories: []model.Factory{
		newModelScopeFactory(),
		newHostsFactory(),
		newInfrastructureFactory(),
		newConfigurationFactory(),
		newKittingFactory(),
		newDistributionFactory(),
		newActivationFactory(),
		newOperationFactory(),
	},

	Regions: model.Regions{
		"local": {
			Id: "us-east-1",
			Az: "us-east-1a",
			Hosts: model.Hosts{
				"ctrl": {
					Scope: model.Scope{Tags: model.Tags{"ctrl"}},
				},
				"terminator": {
					Scope: model.Scope{Tags: model.Tags{"router"}},
				},
				"service": {
					Scope: model.Scope{Tags: model.Tags{"service"}},
				},
			},
		},
		"short": {
			Id: "us-west-1",
			Az: "us-west-1c",
			Hosts: model.Hosts{
				"initiator": {
					Scope: model.Scope{Tags: model.Tags{"router"}},
				},
				"client": {
					Scope: model.Scope{Tags: model.Tags{"client"}},
				},
			},
		},
		"medium": {
			Id: "ap-south-1",
			Az: "ap-south-1a",
			Hosts: model.Hosts{
				"initiator": {
					Scope: model.Scope{Tags: model.Tags{"router"}},
				},
				"client": {
					Scope: model.Scope{Tags: model.Tags{"client"}},
				},
			},
		},
		"long": {
			Id: "ap-southeast-2",
			Az: "ap-southeast-2c",
			Hosts: model.Hosts{
				"initiator": {
					Scope: model.Scope{Tags: model.Tags{"router"}},
				},
				"client": {
					Scope: model.Scope{Tags: model.Tags{"client"}},
				},
			},
		},
	},
}
