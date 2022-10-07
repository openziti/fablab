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

package operation

import (
	"encoding/json"
	"fmt"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Persist() model.OperatingStage {
	return &persist{}
}

func (self *persist) Operate(run model.Run) error {
	if err := self.storeDump(run); err != nil {
		return fmt.Errorf("error storing dump (%w)", err)
	}
	return nil
}

func (self *persist) storeDump(run model.Run) error {
	m := run.GetModel()
	runId := run.GetId()
	dump := m.Dump()

	data, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling dump (%w)", err)
	}

	filename := model.AllocateDump(runId)
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return fmt.Errorf("error creating dump tree [%s] (%w)", filepath.Dir(filename), err)
	}
	if err := ioutil.WriteFile(filename, data, os.ModePerm); err != nil {
		return fmt.Errorf("error writing dump [%s] (%w)", filename, err)
	}

	logrus.Infof("dump saved to [%s]", filename)

	return nil
}

type persist struct {
}
