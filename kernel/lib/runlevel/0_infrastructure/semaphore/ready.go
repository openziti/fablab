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

package semaphore_0

import (
	"errors"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

func Ready(maxWait time.Duration) model.InfrastructureStage {
	return &readyStage{maxWait: maxWait}
}

func (self *readyStage) Express(run model.Run) error {
	m := run.GetModel()

	logrus.Infof("waiting for expressed hosts to be ready (max-wait: %s)", self.maxWait.String())

	start := time.Now()

	for _, r := range m.Regions {
		for _, h := range r.Hosts {
			success := false
			for !success {
				sshConfigFactory := lib.NewSshConfigFactory(h)

				if output, err := lib.RemoteExec(sshConfigFactory, "uptime"); err != nil {
					logrus.Warnf("host not ready [%s] (%v)", h.PublicIp, err)
					if time.Now().Before(start.Add(self.maxWait)) {
						time.Sleep(2 * time.Second)
					} else {
						break
					}
				} else {
					logrus.Infof("%s", strings.Trim(output, " \t\r\n"))
					success = true
				}
			}

			if !success {
				return errors.New("ready check failed. tries exceeded")
			}
		}
	}
	return nil
}

type readyStage struct {
	maxWait time.Duration
}
