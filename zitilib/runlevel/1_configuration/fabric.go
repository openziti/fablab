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

package zitilib_runlevel_1_configuration

import (
	"fmt"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"os"
)

func Fabric() model.ConfigurationStage {
	return &fabric{}
}

func (f *fabric) Configure(run model.Run) error {
	m := run.GetModel()
	if err := generateCa(); err != nil {
		return fmt.Errorf("error generating ca (%s)", err)
	}
	for regionId, region := range m.Regions {
		for hostId, host := range region.Hosts {
			for componentId, component := range host.Components {
				if component.PublicIdentity != "" {
					logrus.Infof("generating public ip identity [%s/%s] on [%s/%s]", componentId, component.PublicIdentity, regionId, hostId)
					if err := generateCert(component.PublicIdentity, host.PublicIp); err != nil {
						return fmt.Errorf("error generating public identity [%s/%s]", componentId, component.PublicIdentity)
					}
				}
				if component.PrivateIdentity != "" {
					logrus.Infof("generating private ip identity [%s/%s] on [%s/%s]", componentId, component.PrivateIdentity, regionId, hostId)
					if err := generateCert(component.PrivateIdentity, host.PrivateIp); err != nil {
						return fmt.Errorf("error generating private identity [%s/%s]", componentId, component.PrivateIdentity)
					}
				}
			}
		}
	}
	return nil
}

func hasExisitingPki() (bool, error) {
	if _, err := os.Stat(model.PkiBuild()); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return true, err
	}
	return true, nil
}

type fabric struct {
}
