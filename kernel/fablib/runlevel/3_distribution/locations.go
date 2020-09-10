/*
	Copyright 2020 NetFoundry, Inc.

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

package distribution

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Locations(hostSpec string, paths ...string) model.DistributionStage {
	return &locations{
		hostSpec: hostSpec,
		paths:    paths,
	}
}

func (self *locations) Distribute(ctx model.RunContext) error {
	m := ctx.GetModel()
	hosts := m.SelectHosts(self.hostSpec)
	for _, host := range hosts {
		ssh := fablib.NewSshConfigFactoryImpl(m, host.PublicIp)
		for _, path := range self.paths {
			mkdir := fmt.Sprintf("mkdir -p %s", path)
			if _, err := fablib.RemoteExec(ssh, mkdir); err == nil {
				logrus.Infof("%s => %s", host.PublicIp, path)
			} else {
				return fmt.Errorf("error creating path [%s] on host [%s] (%w)", path, host.PublicIp, err)
			}
		}
	}
	return nil
}

type locations struct {
	hostSpec string
	paths    []string
}
