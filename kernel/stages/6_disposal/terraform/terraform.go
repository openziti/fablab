/*
	Copyright 2019 Netfoundry, Inc.

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
	"github.com/netfoundry/fablab/kernel/internal"
	"github.com/netfoundry/fablab/model"
	"path/filepath"
)

func Dispose() model.DisposalStage {
	return &terraform{}
}

func (terraform *terraform) Dispose(m *model.Model) error {
	prc := internal.NewProcess("terraform", "destroy", "-auto-approve")
	prc.Cmd.Dir = terraformRun()
	prc.WithTail(internal.StdoutTail)
	if err := prc.Run(); err != nil {
		return fmt.Errorf("error running 'terraform destroy' (%w)", err)
	}
	return nil
}

type terraform struct {
}

func terraformRun() string {
	return filepath.Join(model.ActiveInstancePath(), "tf")
}
