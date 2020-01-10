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

package zitilab_characterization_internet

import (
	"github.com/netfoundry/fablab/kernel/model"
	zitilab_characterization_ziti "github.com/netfoundry/fablab/zitilab/characterization/ziti"
)

func init() {
	model.RegisterModel("zitilab/characterization/internet", Model)
}

// Static model skeleton for zitilab/characterization/internet
//
var Model = &model.Model{
	// Extends zitilab/characterization/ziti
	//
	Parent: zitilab_characterization_ziti.Model,

	Factories: []model.Factory{
		newBindingsFactory(),
	},
}