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

package zitilab

import (
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestIterateScopes(t *testing.T) {
	diamondback.IterateScopes(func(i interface{}, path ...string) {
		if m, ok := i.(*model.Model); ok {
			logrus.Infof("model, tags = %v", m.Tags)
		} else if r, ok := i.(*model.Region); ok {
			logrus.Infof("region %v, tags = %v", path, r.Tags)
		} else if h, ok := i.(*model.Host); ok {
			logrus.Infof("host %v, tags = %v", path, h.Tags)
		} else if c, ok := i.(*model.Component); ok {
			logrus.Infof("component %v, tags = %v", path, c.Tags)
		} else {
			logrus.Infof("%v, s = %p, %s", path, i, reflect.TypeOf(i))
		}
	})
}
