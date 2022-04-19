package model

import (
	"fmt"
	"github.com/openziti/fablab/kernel/lib/timeutil"
	"strings"
	"time"
)

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
