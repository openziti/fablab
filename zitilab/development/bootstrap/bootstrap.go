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

package zitilab_bootstrap

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"os"
)

func (bootstrap *Bootstrap) Bootstrap(m *model.Model) error {
	zitiRoot = os.Getenv("ZITI_ROOT")
	if zitiRoot == "" {
		return fmt.Errorf("please set 'ZITI_ROOT'")
	}
	if fi, err := os.Stat(zitiRoot); err == nil {
		if !fi.IsDir() {
			return fmt.Errorf("invalid 'ZITI_ROOT' (!directory)")
		}
		logrus.Debugf("ZITI_ROOT = [%s]", zitiRoot)
	} else {
		return fmt.Errorf("non-existent 'ZITI_ROOT'")
	}
	return nil
}

type Bootstrap struct{}
