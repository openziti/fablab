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

package component

import (
	"fmt"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
)

func Start(componentSpec string) model.Action {
	return StartInParallel(componentSpec, 1)
}

func StartInParallel(componentSpec string, concurrency int) model.Action {
	return &start{
		componentSpec: componentSpec,
		concurrency:   concurrency,
	}
}

func (start *start) Execute(m *model.Model) error {
	return m.ForEachComponent(start.componentSpec, start.concurrency, func(c *model.Component) error {
		sshConfigFactory := lib.NewSshConfigFactory(c.GetHost())

		if err := lib.LaunchService(sshConfigFactory, c.BinaryName, c.ConfigName); err != nil {
			return fmt.Errorf("error starting component [%s] on [%s] (%s)", c.BinaryName, c.GetHost().PublicIp, err)
		}
		return nil
	})
}

type start struct {
	componentSpec string
	concurrency   int
}
