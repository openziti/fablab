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

package operation

import (
	"encoding/json"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

func Persist() model.OperatingStage {
	return &persist{}
}

func (persist *persist) Operate(m *model.Model) error {
	all := make(map[string]interface{})

	for k, v := range m.Data {
		all["_"+k] = v
	}

	for regionId, region := range m.Regions {
		for k, v := range region.Data {
			all[regionId+"_"+k] = v
		}

		for hostId, host := range region.Hosts {
			for k, v := range host.Data {
				all[regionId+"_"+hostId+"_"+k] = v
			}
		}
	}

	data, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}

	filename := model.AllocateDataset()
	if err := ioutil.WriteFile(filename, data, os.ModePerm); err != nil {
		return err
	}

	logrus.Infof("data saved to [%s]", filename)
	return nil
}

type persist struct {
}
