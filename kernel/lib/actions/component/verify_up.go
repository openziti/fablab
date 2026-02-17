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
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"time"
)

func VerifyUp(componentSpec string, timeout time.Duration) model.Action {
	return VerifyUpInParallel(componentSpec, timeout, 1)
}

func VerifyUpInParallel(componentSpec string, timeout time.Duration, concurrency int) model.Action {
	return &verifyUp{
		componentSpec: componentSpec,
		timeout:       timeout,
		concurrency:   concurrency,
	}
}

type verifyUp struct {
	componentSpec string
	timeout       time.Duration
	concurrency   int
}

func (self *verifyUp) Execute(run model.Run) error {
	return run.GetModel().ForEachComponent(self.componentSpec, self.concurrency, func(c *model.Component) error {
		log := pfxlog.Logger().WithField("componentId", c.Id)
		deadline := time.Now().Add(self.timeout)

		for {
			running, err := c.IsRunning(run)
			if err != nil {
				return err
			}
			if running {
				log.Info("component is running")
				return nil
			}
			if time.Now().After(deadline) {
				log.Error("timed out waiting for component to be running")
				return errors.Errorf("timed out waiting for component [%s] to be running", c.Id)
			}
			log.Info("component not running yet, waiting")
			time.Sleep(500 * time.Millisecond)
		}
	})
}
