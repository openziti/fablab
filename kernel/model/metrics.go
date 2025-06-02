package model

import (
	"fmt"
	"github.com/openziti/fablab/kernel/lib/timeutil"
	"strings"
	"time"
)

type metricSetConverter struct {
	set MetricSet
	err error
}

func (self *metricSetConverter) GetInt64(path ...string) int64 {
	if self.err != nil {
		return 0
	}

	result, ok := self.set.GetInt64Metric(path...)
	if !ok {
		self.err = fmt.Errorf("metric not found: %+v", path)
	}
	return result
}

func (self *metricSetConverter) GetFloat64(path ...string) float64 {
	if self.err != nil {
		return 0
	}

	result, ok := self.set.GetFloat64Metric(path...)
	if !ok {
		self.err = fmt.Errorf("metric not found: %+v", path)
	}
	return result
}

type MetricSet map[string]interface{}

func (s MetricSet) AddGroupedMetric(group, name string, val interface{}) {
	if group == "" || group == name {
		s[name] = val
	} else {
		name = strings.TrimPrefix(name, group+".")
		groupMapVal, found := s[group]
		if !found {
			groupMap := MetricSet{}
			s[group] = groupMap
			groupMap[name] = val
		} else {
			groupMap := groupMapVal.(MetricSet)
			groupMap[name] = val
		}
	}
}

func (s MetricSet) VisitUngroupedMetrics(f func(name string, val interface{})) {
	for k, v := range s {
		if _, ok := v.(MetricSet); !ok {
			f(k, v)
		}
	}
}

func (s MetricSet) VisitGroupedMetrics(f func(name string, group MetricSet)) {
	for k, v := range s {
		if groupSet, ok := v.(MetricSet); ok {
			f(k, groupSet)
		}
	}
}

func (s MetricSet) GetMetric(path ...string) (any, bool) {
	if len(path) == 0 {
		return nil, false
	}

	var current any = s
	for _, k := range path {
		set, ok := current.(MetricSet)
		if !ok {
			return nil, false
		}

		current, ok = set[k]
		if !ok {
			return nil, false
		}
	}

	return current, true
}

func (s MetricSet) GetInt64Metric(path ...string) (int64, bool) {
	result, ok := s.GetMetric(path...)
	if !ok {
		return 0, false
	}
	if intVal, ok := result.(int64); ok {
		return intVal, true
	}
	return 0, false
}

func (s MetricSet) GetFloat64Metric(path ...string) (float64, bool) {
	result, ok := s.GetMetric(path...)
	if !ok {
		return 0, false
	}
	if intVal, ok := result.(float64); ok {
		return intVal, true
	}
	return 0, false
}

func (s MetricSet) AsMeter() (*Meter, error) {
	conv := &metricSetConverter{set: s}
	result := &Meter{
		Count:    conv.GetInt64("count"),
		MeanRate: conv.GetFloat64("mean_rate"),
		M1Rate:   conv.GetFloat64("m1_rate"),
		M5Rate:   conv.GetFloat64("m5_rate"),
		M15Rate:  conv.GetFloat64("m15_rate"),
	}
	if conv.err != nil {
		return nil, conv.err
	}
	return result, nil
}

func (s MetricSet) AsHistogram() (*Histogram, error) {
	conv := &metricSetConverter{set: s}
	result := &Histogram{
		Count:    conv.GetInt64("count"),
		Min:      conv.GetInt64("min"),
		Max:      conv.GetInt64("max"),
		Mean:     conv.GetFloat64("mean"),
		P50:      conv.GetFloat64("p50"),
		P75:      conv.GetFloat64("p75"),
		P95:      conv.GetFloat64("p95"),
		P99:      conv.GetFloat64("p99"),
		StdDev:   conv.GetFloat64("std_dev"),
		Variance: conv.GetFloat64("variance"),
	}
	if conv.err != nil {
		return nil, conv.err
	}
	return result, nil
}

func (s MetricSet) AsTimer() (*Timer, error) {
	conv := &metricSetConverter{set: s}
	result := &Timer{
		Count: conv.GetInt64("count"),

		MeanRate: conv.GetFloat64("mean_rate"),
		M1Rate:   conv.GetFloat64("m1_rate"),
		M5Rate:   conv.GetFloat64("m5_rate"),
		M15Rate:  conv.GetFloat64("m15_rate"),

		Min:      conv.GetInt64("min"),
		Max:      conv.GetInt64("max"),
		Mean:     conv.GetFloat64("mean"),
		P50:      conv.GetFloat64("p50"),
		P75:      conv.GetFloat64("p75"),
		P95:      conv.GetFloat64("p95"),
		P99:      conv.GetFloat64("p99"),
		StdDev:   conv.GetFloat64("std_dev"),
		Variance: conv.GetFloat64("variance"),
	}
	if conv.err != nil {
		return nil, conv.err
	}
	return result, nil
}

type Meter struct {
	Count    int64
	MeanRate float64
	M1Rate   float64
	M5Rate   float64
	M15Rate  float64
}

type Histogram struct {
	Count    int64
	Min      int64
	Max      int64
	Mean     float64
	P50      float64
	P75      float64
	P95      float64
	P99      float64
	StdDev   float64
	Variance float64
}

type Timer struct {
	Count    int64
	MeanRate float64
	M1Rate   float64
	M5Rate   float64
	M15Rate  float64
	Min      int64
	Max      int64
	Mean     float64
	P50      float64
	P75      float64
	P95      float64
	P99      float64
	StdDev   float64
	Variance float64
}

type MetricsEvent struct {
	Timestamp time.Time
	Metrics   MetricSet
	Tags      map[string]string
}

type MetricsHandler interface {
	AcceptHostMetrics(host *Host, event *MetricsEvent)
}

type DataMetricsWriter struct {
}

func (DataMetricsWriter) AcceptHostMetrics(host *Host, event *MetricsEvent) {
	var metricsSlice []map[string]interface{}
	val, found := host.Data["metrics"]
	if found {
		metricsSlice = val.([]map[string]interface{})
	}
	metricsMap := map[string]interface{}{}
	metricsMap["timestamp_ms"] = fmt.Sprintf("%v", timeutil.TimeToMilliseconds(event.Timestamp))
	for name, val := range event.Metrics {
		metricsMap[name] = val
	}
	host.Data["metrics"] = append(metricsSlice, metricsMap)
}

type StdOutMetricsWriter struct {
}

func (StdOutMetricsWriter) AcceptHostMetrics(host *Host, event *MetricsEvent) {
	fmt.Printf("metrics event - host %v at timestamp: %v\n", host.GetId(), event.Timestamp)
	for k, v := range event.Metrics {
		fmt.Printf("\t%v = %v\n", k, v)
	}
}
