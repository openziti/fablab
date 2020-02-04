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

package zitilib_characterization_ziti

import "github.com/netfoundry/fablab/kernel/model"

func newModelScopeFactory() model.Factory {
	return &modelScopeFactory{}
}

func (f *modelScopeFactory) Build(m *model.Model) error {
	m.Variables = model.Variables{
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
	}
	return nil
}

type modelScopeFactory struct{}
