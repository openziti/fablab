/*
	(c) Copyright NetFoundry Inc. Inc.

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

// DefaultParallelism is the default number of concurrent terraform operations.
const DefaultParallelism = 5

func Dispose() model.Stage {
	return &terraform{
		Parallelism: DefaultParallelism,
	}
}

func (t *terraform) Execute(model.Run) error {
	args := []string{"destroy", "-auto-approve"}
	if t.Parallelism > 0 {
		args = append(args, fmt.Sprintf("-parallelism=%d", t.Parallelism))
	}
	prc := lib.NewProcess("terraform", args...)
	prc.Cmd.Dir = terraformRun()
	prc.WithTail(lib.StdoutTail)
	if err := prc.Run(); err != nil {
		return fmt.Errorf("error running 'terraform destroy' (%w)", err)
	}
	return nil
}

type terraform struct {
	Parallelism int
}

func terraformRun() string {
	return filepath.Join(model.BuildPath(), "tf")
}
