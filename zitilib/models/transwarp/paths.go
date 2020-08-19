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

package transwarp

import (
	"fmt"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"sync"
)

var dilithiumRoot string
var bootstrapOnce = sync.Once{}

func DilithiumRoot() string {
	return dilithiumRoot
}

func DilithiumEtc() string {
	return filepath.Join(dilithiumRoot, "etc")
}

func (self *bootstrap) Bootstrap(_ *model.Model) error {
	var err error
	bootstrapOnce.Do(func() {
		dilithiumRoot = os.Getenv("DILITHIUM_ROOT")
		if dilithiumRoot == "" {
			err = errors.New("please set 'DILITHIUM_ROOT'")
			return
		}
		if fi, err := os.Stat(dilithiumRoot); err == nil {
			if !fi.IsDir() {
				err = fmt.Errorf("invalid 'DILITHIUM_ROOT' (!directory)")
				return
			}
		} else {
			err = errors.New("non-existent 'DILITHIUM_ROOT'")
			return
		}
	})
	return err
}

type bootstrap struct{}