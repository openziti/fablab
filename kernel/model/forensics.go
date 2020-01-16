/*
	Copyright 2019 NetFoundry, Inc.

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
	"github.com/netfoundry/ziti-foundation/util/info"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func AllocateDataset() string {
	return fmt.Sprintf("%s/data-%d.json", ActiveInstancePath(), info.NowInMilliseconds())
}

func ListDatasets() ([]string, error) {
	files, err := ioutil.ReadDir(ActiveInstancePath())
	if err != nil {
		return nil, fmt.Errorf("unable to list datasets (%w)", err)
	}

	var datasets []string
	for _, file := range files {
		if file.Mode().IsRegular() && strings.HasPrefix(file.Name(), "data-") && strings.HasSuffix(file.Name(), ".json") {
			datasets = append(datasets, filepath.Join(ActiveInstancePath(), file.Name()))
		}
	}

	return datasets, nil
}
