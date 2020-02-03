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
	"github.com/oliveagle/jsonpath"
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
	if dumppaths, err := model.ListDumps(); err == nil {
		for i, dumppath := range dumppaths {
			data, err := ioutil.ReadFile(dumppath)
			if err != nil {
				return fmt.Errorf("unable to read dump [%s] (%w)", dumppath, err)
			}

			dump := &model.Dump{}
			if err := json.Unmarshal(data, dump); err != nil {
				return fmt.Errorf("error unmarshaling dump (%w)", err)
			}

			reportData, err := report.buildReportData(data, []string{"short", "medium", "long"})
			if err != nil {
				return fmt.Errorf("error building report data (%w)", err)
			}
			reportData.Dump = dump

			tPath := filepath.Join(model.FablabRoot(), "zitilib/characterization/reporting/templates/index.html")

			reportPath := filepath.Join(model.ActiveInstancePath(), fmt.Sprintf("reports/%d.html", i))
			if err := os.MkdirAll(filepath.Dir(reportPath), os.ModePerm); err != nil {
				return fmt.Errorf("error creating report path [%s] (%w)", reportPath, err)
			}
			if err := report.renderTemplate(tPath, reportPath, reportData); err != nil {
				return fmt.Errorf("unable to render template (%w)", err)
			}

			logrus.Infof("[%s] => [%s]", dumppath, reportPath)
		}

		return nil

	} else {
		return err
	}
}

func (report *report) buildReportData(data []byte, regionKeys []string) (*ReportData, error) {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("error unmarshalling json data (%w)", err)
	}

	reportData := &ReportData{
		RegionKeys: regionKeys,
		RegionData: make(map[string]*ReportRegionData),
	}

	for _, regionKey := range regionKeys {
		regionData := &ReportRegionData{}

		iperfSummary, err := report.getIperfSummary(jsonData, fmt.Sprintf("$.regions.%s.hosts.client.scope.data.iperf_ziti_metrics", regionKey))
		if err != nil {
			return nil, fmt.Errorf("error getting ziti iperf summary (%w)", err)
		}
		regionData.Ziti.IPerf = iperfSummary

		iperfSummary, err = report.getIperfSummary(jsonData, fmt.Sprintf("$.regions.%s.hosts.client.scope.data.iperf_internet_metrics", regionKey))
		if err != nil {
			return nil, fmt.Errorf("error getting internet iperf summary (%w)", err)
		}
		regionData.Internet.IPerf = iperfSummary

		iperfUdpSummary, err := report.getIperfUdpSummary(jsonData, fmt.Sprintf("$.regions.%s.hosts.client.scope.data.iperf_udp_ziti_1m_metrics", regionKey))
		if err != nil {
			return nil, fmt.Errorf("error getting ziti iperf udp summary (%w)", err)
		}
		regionData.Ziti.IPerfUdp = iperfUdpSummary

		iperfUdpSummary, err = report.getIperfUdpSummary(jsonData, fmt.Sprintf("$.regions.%s.hosts.client.scope.data.iperf_udp_internet_1m_metrics", regionKey))
		if err != nil {
			return nil, fmt.Errorf("error getting internet iperf udp summary (%w)", err)
		}
		regionData.Internet.IPerfUdp = iperfUdpSummary

		reportData.RegionData[regionKey] = regionData
	}

	return reportData, nil
}

func (report *report) getIperfSummary(jsonData interface{}, path string) (*model.IperfSummary, error) {
	compiled, err := jsonpath.Compile(path)
	if err != nil {
		return nil, fmt.Errorf("error compiling json path [%s] (%w)", path, err)
	}

	res, err := compiled.Lookup(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error querying json path [%s] (%w)", path, err)
	}

	data, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json (%w)", err)
	}

	summary := &model.IperfSummary{}
	if err := json.Unmarshal(data, summary); err != nil {
		return nil, fmt.Errorf("error unmarshaling iperf summary (%w)", err)
	}

	return summary, nil
}
func (report *report) getIperfUdpSummary(jsonData interface{}, path string) (*model.IperfUdpSummary, error) {
	compiled, err := jsonpath.Compile(path)
	if err != nil {
		return nil, fmt.Errorf("error compiling json path [%s] (%w)", path, err)
	}

	res, err := compiled.Lookup(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error querying json path [%s] (%w)", path, err)
	}

	data, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json (%w)", err)
	}

	summary := &model.IperfUdpSummary{}
	if err := json.Unmarshal(data, summary); err != nil {
		return nil, fmt.Errorf("error unmarshaling iperf udp summary (%w)", err)
	}

	return summary, nil
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
	Dump       *model.Dump
	RegionKeys []string
	RegionData map[string]*ReportRegionData
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
