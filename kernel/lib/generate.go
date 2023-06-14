/*
	Copyright 2020 NetFoundry Inc.

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

package lib

import (
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/resources"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/fs"
	"path/filepath"
)

func GenerateConfigForComponent(c *model.Component, fs fs.FS, src, configName string, r model.Run) error {
	logrus.Debugf("generating configuration for component [%s/%s/%s]", c.Region().Id, c.Host.Id, c.Id)

	if fs == nil {
		fs = c.GetModel().GetResource(resources.Configs)
	}

	dst := filepath.Join(r.GetConfigDir(), configName)
	err := RenderTemplateFS(fs, src, dst, c.GetModel(), &struct {
		RegionId  string
		HostId    string
		Host      *model.Host
		Component *model.Component
		Model     *model.Model
	}{
		RegionId:  c.GetRegion().Id,
		HostId:    c.GetHost().GetId(),
		Host:      c.GetHost(),
		Component: c,
		Model:     c.GetModel(),
	})
	if err != nil {
		return errors.Wrapf(err, "error rendering template [%s]", src)
	}

	logrus.Infof("config [%s] => [%s]", src, configName)

	return nil
}
