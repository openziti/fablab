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

package reporting

import (
	"encoding/json"
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/ziti-foundation/util/info"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func Report() model.Action {
	return &report{}
}

func (report *report) Execute(m *model.Model) error {
	tData := &templateData{}

	if datasets, err := model.ListDatasets(); err == nil {
		for _, dataset := range datasets {
			data, err := ioutil.ReadFile(dataset)
			if err != nil {
				return fmt.Errorf("unable to read dataset [%s] (%w)", dataset, err)
			}

			datamap := make(map[string]interface{})
			if err := json.Unmarshal(data, &datamap); err != nil {
				return fmt.Errorf("error unmarshalling dataset [%s] (%w)", dataset, err)
			}

			metricSets := make([]string, 0)
			for k := range datamap {
				metricSets = append(metricSets, k)
			}
			tData.MetricSets = append(tData.MetricSets, metricSets)

			tData.Datasets = append(tData.Datasets, datamap)

			logrus.Infof("dataset = [%s] (%s)", dataset, info.ByteCount(int64(len(data))))
		}

		tPath := filepath.Join(model.FablabRoot(), "zitilab/characterization/reporting/templates/index.html")
		if err := report.renderTemplate(tPath, "index.html", tData); err != nil {
			return fmt.Errorf("unable to render template (%w)", err)
		}
		return nil

	} else {
		return err
	}
}

func (report *report) renderTemplate(src, dst string, tData *templateData) error {
	tSrc, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("error reading template [%s] (%w)", src, err)
	}

	t, err := template.New("report").Funcs(report.templateFuncs()).Parse(string(tSrc))
	if err != nil {
		return fmt.Errorf("error parsing template [%s] (%w)", src, err)
	}

	dstF, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating output file [%s] (%w)", dst, err)
	}
	defer func() { _ = dstF.Close() }()

	if err := t.Execute(dstF, tData); err != nil {
		return fmt.Errorf("error rendering template [%s] (%w)", src, err)
	}

	logrus.Infof("wrote report to => [%s]", dst)

	return nil
}

func (report *report) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"json": func(i interface{}) string {
			data, err := json.MarshalIndent(i, "", "  ")
			if err != nil {
				logrus.Fatalf("error marshaling json (%w)", err)
			}
			return string(data)
		},
		"isIperfMetrics": func(metricName string) bool {
			return strings.Contains(metricName, "_iperf_")
		},
	}
}

type report struct{}

type templateData struct {
	Datasets   []interface{}
	MetricSets [][]string
}
