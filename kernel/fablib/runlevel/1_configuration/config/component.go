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

package config

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

func Component() model.ConfigurationStage {
	return &componentConfig{}
}

func (componentConfig *componentConfig) Configure(m *model.Model) error {
	for regionId, r := range m.Regions {
		for hostId, h := range r.Hosts {
			if err := componentConfig.generateComponentForHost(regionId, hostId, m, h); err != nil {
				return fmt.Errorf("error generating config for host [%s/%s] (%s)", regionId, hostId, err)
			}
		}
	}
	return nil
}

func (componentConfig *componentConfig) generateComponentForHost(regionId, hostId string, m *model.Model, h *model.Host) error {
	for componentId, c := range h.Components {
		if c.ScriptSrc != "" && c.ScriptName != "" {
			if err := componentConfig.generateScriptForComponent(regionId, hostId, componentId, m, h, c); err != nil {
				return err
			}
		}
		if c.ConfigSrc != "" && c.ConfigName != "" {
			if err := componentConfig.generateConfigForComponent(regionId, hostId, componentId, m, h, c); err != nil {
				return err
			}
		}
	}
	return nil
}

func (componentConfig *componentConfig) generateScriptForComponent(regionId, hostId, componentId string, m *model.Model, h *model.Host, c *model.Component) error {
	logrus.Debugf("generating script for component [%s/%s/%s]", regionId, hostId, componentId)

	src := filepath.Join(model.ScriptSrc(), c.ScriptSrc)
	dst := filepath.Join(model.ScriptBuild(), c.ScriptName)
	err := fablib.RenderTemplate(src, dst, m, &templateModel{
		RegionId:  regionId,
		HostId:    hostId,
		Host:      h,
		Component: c,
		Model:     m,
	})
	if err != nil {
		return fmt.Errorf("error rendering template (%w)", err)
	}

	logrus.Infof("script [%s] => [%s]", c.ScriptSrc, c.ScriptName)

	return nil
}

func (componentConfig *componentConfig) generateConfigForComponent(regionId, hostId, componentId string, m *model.Model, h *model.Host, c *model.Component) error {
	logrus.Debugf("generating configuration for component [%s/%s/%s]", regionId, hostId, componentId)

	src := filepath.Join(model.ConfigSrc(), c.ConfigSrc)
	dst := filepath.Join(model.ConfigBuild(), c.ConfigName)
	err := fablib.RenderTemplate(src, dst, m, &templateModel{
		RegionId:  regionId,
		HostId:    hostId,
		Host:      h,
		Component: c,
		Model:     m,
	})
	if err != nil {
		return fmt.Errorf("error rendering template (%w)", err)
	}

	logrus.Infof("config [%s] => [%s]", c.ConfigSrc, c.ConfigName)

	return nil
}

type componentConfig struct {
}
