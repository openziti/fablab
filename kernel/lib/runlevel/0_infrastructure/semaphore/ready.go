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

package semaphore_0

import (
	"errors"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

func Ready(maxWait time.Duration) model.Stage {
	return &ReadyStage{MaxWait: maxWait}
}

func (self *ReadyStage) Execute(run model.Run) error {
	logrus.Infof("waiting for expressed hosts to be ready (max-wait: %s)", self.MaxWait.String())

	start := time.Now()

	return run.GetModel().ForEachHost("*", 20, func(host *model.Host) error {
		for {
			output, err := host.ExecLogged("uptime")
			if err == nil {
				logrus.Infof("%s", strings.Trim(output, " \t\r\n"))
				return nil
			}

			logrus.Warnf("host not ready [%s] (%v)", host.PublicIp, err)
			if time.Now().Before(start.Add(self.MaxWait)) {
				time.Sleep(2 * time.Second)
			} else {
				return errors.New("ready check failed. tries exceeded")
			}
		}
	})
}

type ReadyStage struct {
	MaxWait time.Duration
}
