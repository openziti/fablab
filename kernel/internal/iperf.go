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

package internal

import (
	"encoding/json"
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/oliveagle/jsonpath"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

func SummarizeIperf(data []byte) ([]model.IperfSummary, error) {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("error unmarshaling [%w]", err)
	}

	startPath, err := jsonpath.Compile("$.start.timestamp.timesecs")
	if err != nil {
		return nil, fmt.Errorf("error in compiling json timestamp path (%w)", err)
	}
	res, err := startPath.Lookup(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error in start path lookup (%w)", err)
	}

	start := time.Unix(int64(res.(float64)), 0)

	sumPath, err := jsonpath.Compile("$.intervals.sum")
	if err != nil {
		return nil, fmt.Errorf("error compiling json sum path (%w)", err)
	}
	res, err = sumPath.Lookup(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error in json lookup [%w]", err)
	}

	summary := make([]model.IperfSummary, 0)
	for _, value := range res.([]interface{}) {
		sum, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot cast 'sum' [%s]", reflect.TypeOf(value))
		}

		sumStart := sum["start"].(float64)
		timestamp := start.Add(time.Duration(int64(sumStart*1000.0)) * time.Millisecond)
		bitsPerSecond := sum["bits_per_second"].(float64)
		summary = append(summary, model.IperfSummary{
			TimestampMs:   TimeToMilliseconds(timestamp),
			BitsPerSecond: bitsPerSecond,
		})

		logrus.Debugf("start = [%f], bits_per_second = [%f]", sum["start"].(float64), sum["bits_per_second"].(float64))
	}
	return summary, nil
}
