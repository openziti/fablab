package internal

import (
	"encoding/json"
	"fmt"
	"github.com/oliveagle/jsonpath"
	"github.com/sirupsen/logrus"
	"reflect"
)

func SummarizeIperf(data []byte) ([]IperfSummary, error) {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("error unmarshaling [%w]", err)
	}

	bpsPath, err := jsonpath.Compile("$.intervals.sum")
	if err != nil {
		return nil, fmt.Errorf("error compiling json path [%w]", err)
	}

	res, err := bpsPath.Lookup(jsonData)
	if err != nil {
		return nil, fmt.Errorf("error in json lookup [%w]", err)
	}

	summary := make([]IperfSummary, 0)
	for _, value := range res.([]interface{}) {
		sum, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot cast 'sum' [%s]", reflect.TypeOf(value))
		}
		start := sum["start"].(float64)
		bitsPerSecond := sum["bits_per_second"].(float64)
		summary = append(summary, IperfSummary{Start: start, BitsPerSecond: bitsPerSecond})
		logrus.Infof("start = [%f], bits_per_second = [%f]", sum["start"].(float64), sum["bits_per_second"].(float64))
	}
	return summary, nil
}

type IperfSummary struct {
	Start         float64 `json:"start"`
	BitsPerSecond float64 `json:"bits_per_second"`
}
