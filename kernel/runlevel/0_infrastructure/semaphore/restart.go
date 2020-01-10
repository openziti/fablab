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

package semaphore_0

import (
	"errors"
	"github.com/netfoundry/fablab/kernel/internal"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

func Restart(preDelay time.Duration) model.InfrastructureStage {
	return &restartStage{preDelay: preDelay}
}

func (restartStage *restartStage) Express(m *model.Model, l *model.Label) error {
	logrus.Infof("waiting for expressed hosts to restart (pre-delay: %s)", restartStage.preDelay.String())
	time.Sleep(restartStage.preDelay)

	sshUsername := m.MustVariable("credentials", "ssh", "username").(string)

	logrus.Infof("starting restart checks")
	for _, r := range m.Regions {
		for _, h := range r.Hosts {
			success := false
			for tries := 0; tries < 5; tries++ {
				if output, err := internal.RemoteExec(sshUsername, h.PublicIp, "uptime"); err != nil {
					logrus.Warnf("host not restarted [%s] (%w)", h.PublicIp, err)
					time.Sleep(10 * time.Second)
				} else {
					logrus.Infof("%s", strings.Trim(output, " \t\r\n"))
					success = true
					break
				}
			}
			if !success {
				return errors.New("restart check failed. tries exceeded")
			}
		}
	}
	return nil
}

type restartStage struct {
	preDelay time.Duration
}
