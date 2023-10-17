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

package component

import (
	"github.com/openziti/fablab/kernel/model"
)

func Stop(componentSpec string) model.Action {
	return StopInParallel(componentSpec, 1)
}

func StopInParallel(componentSpec string, concurrency int) model.Action {
	return &stop{
		componentSpec: componentSpec,
		concurrency:   concurrency,
	}
}

func StopInParallelHostExclusive(componentSpec string, concurrency int) model.Action {
	return &stopByHost{
		componentSpec: componentSpec,
		concurrency:   concurrency,
	}
}

func (stop *stop) Execute(run model.Run) error {
	return run.GetModel().ForEachComponent(stop.componentSpec, stop.concurrency, func(c *model.Component) error {
		if c.Type != nil {
			return c.Type.Stop(run, c)
		}
		return nil
	})
}

type stop struct {
	componentSpec string
	concurrency   int
}

type stopByHost struct {
	componentSpec string
	concurrency   int
}

func (stop *stopByHost) Execute(run model.Run) error {
	return run.GetModel().ForEachComponent(stop.componentSpec, stop.concurrency, func(c *model.Component) error {
		return c.Host.DoExclusiveFallible(func() error {
			if c.Type != nil {
				return c.Type.Stop(run, c)
			}
			return nil
		})
	})
}
