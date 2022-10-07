/*
	Copyright 2019 NetFoundry Inc.

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

package model

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

func loadBindings() error {
	var data []byte
	var err error
	data, err = ioutil.ReadFile(bindingsYml())
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("no bindings [%s]", bindingsYml())
		} else {
			return fmt.Errorf("error reading bindings [%s] (%w)", bindingsYml(), err)
		}
	}

	bindings = make(map[string]interface{})
	if err := yaml.Unmarshal(data, &bindings); err != nil {
		return fmt.Errorf("error unmarshalling bindings [%s] (%w)", bindingsYml(), err)
	}
	return nil
}

func bindingsYml() string {
	return filepath.Join(configRoot(), "bindings.yml")
}
