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
	"text/template"
)

func Report() model.Action {
	return &report{}
}

func (report *report) Execute(m *model.Model) error {
	if datasets, err := model.ListDatasets(); err == nil {
		tData := &ReportData{Regions: make(map[string]*ReportRegionData)}

		for i, dataset := range datasets {
			data, err := ioutil.ReadFile(dataset)
			if err != nil {
				return fmt.Errorf("unable to read dataset [%s] (%w)", dataset, err)
			}

			datamap := make(map[string]interface{})
			if err := json.Unmarshal(data, &datamap); err != nil {
				return fmt.Errorf("error unmarshalling dataset [%s] (%w)", dataset, err)
			}

			tData.RegionKeys = []string{"short", "medium", "long"}
			for _, regionPrefix := range tData.RegionKeys {
				regionData := &ReportRegionData{}

				key := fmt.Sprintf("%s_client_iperf_ziti_metrics", regionPrefix)
				if value, found := datamap[key]; found {
					summary, err := report.toIperfSummary(value)
					if err != nil {
						return fmt.Errorf("error conforming [%s] (%w)", key, err)
					}
					regionData.Ziti.IPerf = summary
				} else {
					return fmt.Errorf("missing [%s]", key)
				}

				key = fmt.Sprintf("%s_client_iperf_internet_metrics", regionPrefix)
				if value, found := datamap[key]; found {
					summary, err := report.toIperfSummary(value)
					if err != nil {
						return fmt.Errorf("error conforming [%s] (%w)", key, err)
					}
					regionData.Internet.IPerf = summary
				} else {
					return fmt.Errorf("missing [%s]", key)
				}

				key = fmt.Sprintf("%s_client_iperf_udp_ziti_1m_metrics", regionPrefix)
				if value, found := datamap[key]; found {
					summary, err := report.toIperfUdpSummary(value)
					if err != nil {
						return fmt.Errorf("error conforming [%s] (%w)", key, err)
					}
					regionData.Ziti.IPerfUdp = summary
				} else {
					return fmt.Errorf("missing [%s]", key)
				}

				key = fmt.Sprintf("%s_client_iperf_udp_internet_1m_metrics", regionPrefix)
				if value, found := datamap[key]; found {
					summary, err := report.toIperfUdpSummary(value)
					if err != nil {
						return fmt.Errorf("error conforming [%s] (%w)", key, err)
					}
					regionData.Internet.IPerfUdp = summary
				} else {
					return fmt.Errorf("missing [%s]", key)
				}

				tData.Regions[regionPrefix] = regionData
			}

			tPath := filepath.Join(model.FablabRoot(), "zitilab/characterization/reporting/templates/index.html")

			reportPath := filepath.Join(model.ActiveInstancePath(), fmt.Sprintf("reports/%d.html", i))
			if err := os.MkdirAll(filepath.Dir(reportPath), os.ModePerm); err != nil {
				return fmt.Errorf("error creating report path [%s] (%w)", reportPath, err)
			}
			if err := report.renderTemplate(tPath, reportPath, tData); err != nil {
				return fmt.Errorf("unable to render template (%w)", err)
			}

			logrus.Infof("[%s] => [%s]", dataset, reportPath)
		}

		return nil

	} else {
		return err
	}
}

func (report *report) toIperfSummary(v interface{}) (*model.IperfSummary, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json (%w)", err)
	}

	iperfSummary := &model.IperfSummary{}
	if err := json.Unmarshal(data, iperfSummary); err != nil {
		return nil, fmt.Errorf("error unmarshaling iperf summary (%w)", err)
	}

	return iperfSummary, nil
}

func (report *report) toIperfUdpSummary(v interface{}) (*model.IperfUdpSummary, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json (%w)", err)
	}

	iperfUdpSummary := &model.IperfUdpSummary{}
	if err := json.Unmarshal(data, iperfUdpSummary); err != nil {
		return nil, fmt.Errorf("error unmarshaling iperf udp summary (%w)", err)
	}

	return iperfUdpSummary, nil
}

func (report *report) renderTemplate(src, dst string, data *ReportData) error {
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

	if err := t.Execute(dstF, data); err != nil {
		return fmt.Errorf("error rendering template [%s] (%w)", src, err)
	}

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
		"dataRate": func(value float64) string {
			return info.ByteCount(int64(value))
		},
	}
}

type report struct{}

type ReportData struct {
	RegionKeys []string
	Regions    map[string]*ReportRegionData
}

type ReportRegionData struct {
	Ziti struct {
		IPerf    *model.IperfSummary
		IPerfUdp *model.IperfUdpSummary
	}
	Internet struct {
		IPerf    *model.IperfSummary
		IPerfUdp *model.IperfUdpSummary
	}
}
