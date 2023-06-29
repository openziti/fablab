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

package terraform

import (
	"fmt"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"path/filepath"
)

func Dispose() model.Stage {
	return &terraform{}
}

func (terraform *terraform) Execute(model.Run) error {
	prc := lib.NewProcess("terraform", "destroy", "-auto-approve")
	prc.Cmd.Dir = terraformRun()
	prc.WithTail(lib.StdoutTail)
	if err := prc.Run(); err != nil {
		return fmt.Errorf("error running 'terraform destroy' (%w)", err)
	}
	return nil
}

type terraform struct {
}

func terraformRun() string {
	return filepath.Join(model.BuildPath(), "tf")
}
