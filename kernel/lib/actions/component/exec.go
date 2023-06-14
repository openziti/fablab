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
	"github.com/pkg/errors"
)

func Exec[T model.ComponentType](componentSpec string, strategyAction func(strategy T, run model.Run, component *model.Component) error) model.Action {
	return ExecInParallel(componentSpec, 1, strategyAction)
}

func ExecInParallel[T model.ComponentType](componentSpec string, concurrency int, strategyAction func(strategy T, run model.Run, component *model.Component) error) model.Action {
	return &exec{
		componentSpec: componentSpec,
		concurrency:   concurrency,
		f: func(run model.Run, c *model.Component) error {
			return Dispatch(run, c, strategyAction)
		},
	}
}

func ExecIfApplies[T model.ComponentType](componentSpec string, strategyAction func(strategy T, run model.Run, component *model.Component) error) model.Action {
	return ExecIfAppliesInParallel(componentSpec, 1, strategyAction)
}

func Dispatch[T model.ComponentType](run model.Run, component *model.Component, strategyAction func(strategy T, run model.Run, component *model.Component) error) error {
	if component.Type == nil {
		return errors.Errorf("component %v has no strategy", component.Id)
	}
	typedStrategy, ok := component.Type.(T)
	if !ok {
		exampleInstance := new(T)
		return errors.Errorf("component %v has has the wrong strategy type, has %T, expected %T", component.Id, component.Type, exampleInstance)
	}
	return strategyAction(typedStrategy, run, component)
}

func DispatchIfApplies[T model.ComponentType](run model.Run, component *model.Component, strategyAction func(strategy T, run model.Run, component *model.Component) error) error {
	if component.Type == nil {
		return nil
	}
	typedStrategy, ok := component.Type.(T)
	if !ok {
		return nil
	}
	return strategyAction(typedStrategy, run, component)
}

func ExecIfAppliesInParallel[T model.ComponentType](componentSpec string, concurrency int, strategyAction func(strategy T, run model.Run, component *model.Component) error) model.Action {
	return &exec{
		componentSpec: componentSpec,
		concurrency:   concurrency,
		f: func(run model.Run, c *model.Component) error {
			return DispatchIfApplies(run, c, strategyAction)
		},
	}
}

type exec struct {
	componentSpec string
	concurrency   int
	f             func(run model.Run, c *model.Component) error
}

func (self *exec) Execute(run model.Run) error {
	return run.GetModel().ForEachComponent(self.componentSpec, self.concurrency, func(c *model.Component) error {
		return self.f(run, c)
	})
}
