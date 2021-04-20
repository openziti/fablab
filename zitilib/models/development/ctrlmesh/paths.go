/*
	Copyright (c) NetFoundry, Inc.

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

package ctrlmesh

import (
	"fmt"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"sync"
)

var ctrlmeshRoot string
var bootstrapOnce = sync.Once{}

func CtrlmeshRoot() string {
	return ctrlmeshRoot
}

func CtrlmeshBinaries() string {
	return filepath.Join(CtrlmeshRoot(), "bin")
}

func (_ *bootstrap) Bootstrap(_ *model.Model) error {
	var err error
	bootstrapOnce.Do(func() {
		ctrlmeshRoot = os.Getenv("CTRLMESH_ROOT")
		if ctrlmeshRoot == "" {
			err = errors.New("please set 'CTRLMESH_ROOT'")
			return
		}
		if fi, err := os.Stat(ctrlmeshRoot); err == nil {
			if !fi.IsDir() {
				err = fmt.Errorf("invalid 'CTRLMESH_ROOT' (!directory)")
				return
			}
		} else {
			err = errors.New("non-existent 'CTRLMESH_ROOT'")
			return
		}
	})
	return err
}

type bootstrap struct{}