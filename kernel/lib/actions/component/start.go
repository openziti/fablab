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

package component

import (
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

func (start *start) Execute(run model.Run) error {
	return run.GetModel().ForEachComponent(start.componentSpec, start.concurrency, func(c *model.Component) error {
		if startable, ok := c.Type.(model.ServerComponent); ok {
			return startable.Start(run, c)
		}
		return nil
	})
}

type start struct {
	componentSpec string
	concurrency   int
}
