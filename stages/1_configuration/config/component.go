/*
	Copyright 2019 Netfoundry, Inc.

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

package config

import (
	"bitbucket.org/netfoundry/fablab/kernel"
	"bitbucket.org/netfoundry/fablab/kernel/lib"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

func Component() kernel.ConfigurationStage {
	return &componentConfig{}
}

func (c *componentConfig) Configure(m *kernel.Model) error {
	for regionId, region := range m.Regions {
		for hostId, host := range region.Hosts {
			if err := c.generateConfigForHost(regionId, hostId, host, m); err != nil {
				return fmt.Errorf("error generating config for host [%s/%s] (%s)", regionId, hostId, err)
			}
		}
	}
	return nil
}

func (c *componentConfig) generateConfigForHost(regionId, hostId string, host *kernel.Host, model *kernel.Model) error {
	for componentName, component := range host.Components {
		logrus.Debugf("generating configuration for component [%s/%s/%s]", regionId, hostId, componentName)

		tPath := filepath.Join(kernel.ConfigSrc(), component.ConfigSrc)
		tData, err := ioutil.ReadFile(tPath)
		if err != nil {
			return fmt.Errorf("error reading template [%s] (%w)", tPath, err)
		}

		t, err := template.New("config").Funcs(lib.TemplateFuncMap(model)).Parse(string(tData))
		if err != nil {
			return fmt.Errorf("error parsing template [%s] (%w)", tPath, err)
		}

		outputPath := filepath.Join(kernel.ConfigBuild(), component.ConfigName)
		if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
			return fmt.Errorf("error creating directories [%s] (%w)", outputPath, err)
		}

		outputF, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating config [%s] (%w)", outputPath, err)
		}
		defer func() { _ = outputF.Close() }()

		err = t.Execute(outputF, &templateModel{
			RegionId:  regionId,
			HostId:    hostId,
			Host:      host,
			Component: component,
			Model:     model,
		})
		if err != nil {
			return fmt.Errorf("error rendering template [%s] (%w)", outputPath, err)
		}

		logrus.Infof("config [%s] => [%s]", component.ConfigSrc, component.ConfigName)
	}

	return nil
}

type componentConfig struct {
}
