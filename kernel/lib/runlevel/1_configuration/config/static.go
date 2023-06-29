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

package config

import (
	"fmt"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fablab/resources"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
)

func Static(configs []StaticConfig) model.Stage {
	return &staticConfig{configs: configs}
}

func (staticConfig *staticConfig) Execute(run model.Run) error {
	m := run.GetModel()
	configResource := run.GetModel().GetResource(resources.Configs)
	for _, config := range staticConfig.configs {
		logrus.Debugf("generating static configuration [%s] => [%s]", config.Src, config.Name)

		tData, err := fs.ReadFile(configResource, config.Src)
		if err != nil {
			return fmt.Errorf("error reading template [%s] (%w)", config.Src, err)
		}

		t, err := template.New("config").Funcs(lib.TemplateFuncMap(m)).Parse(string(tData))
		if err != nil {
			return fmt.Errorf("error parsing template [%s] (%w)", config.Src, err)
		}

		outputPath := filepath.Join(model.ConfigBuild(), config.Name)
		if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
			return fmt.Errorf("error creating directories [%s] (%w)", filepath.Dir(outputPath), err)
		}

		outputF, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating config [%s] (%w)", outputPath, err)
		}
		defer func() { _ = outputF.Close() }()

		err = t.Execute(outputF, &templateModel{
			Model: m,
		})
		if err != nil {
			return fmt.Errorf("error rendering template [%s] (%w)", outputPath, err)
		}

		logrus.Infof("[%s] => [%s]", config.Src, outputPath)
	}

	return nil
}

type StaticConfig struct {
	Src  string
	Name string
}

type staticConfig struct {
	configs []StaticConfig
}
