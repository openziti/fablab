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

package model

import (
	"github.com/openziti/fablab/kernel/lib/timeutil"
)

type HostSummary struct {
	Cpu     []*CpuTimeslice     `json:"cpu,omitempty"`
	Memory  []*MemoryTimeslice  `json:"memory,omitempty"`
	Process []*ProcessTimeslice `json:"process,omitempty"`
}

func (hs *HostSummary) ToMetricsEvents() (events []*MetricsEvent) {
	for _, e := range hs.Cpu {
		events = append(events, e.toMetricsEvent())
	}
	for _, e := range hs.Memory {
		events = append(events, e.toMetricsEvent())
	}
	for _, e := range hs.Process {
		events = append(events, e.toMetricsEvent())
	}
	return events
}

type CpuTimeslice struct {
	TimestampMs   int64   `json:"timestamp_ms"`
	PercentUser   float64 `json:"percent_user"`
	PercentNice   float64 `json:"percent_nice"`
	PercentSystem float64 `json:"percent_system"`
	PercentIowait float64 `json:"percent_iowait"`
	PercentSteal  float64 `json:"percent_steal"`
	PercentIdle   float64 `json:"percent_idle"`
}

func (ts *CpuTimeslice) toMetricsEvent() *MetricsEvent {
	event := &MetricsEvent{
		Timestamp: timeutil.MillisecondsToTime(ts.TimestampMs),
		Metrics:   MetricSet{},
	}

	event.Metrics["percent_user"] = ts.PercentUser
	event.Metrics["percent_nice"] = ts.PercentNice
	event.Metrics["percent_system"] = ts.PercentSystem
	event.Metrics["percent_iowait"] = ts.PercentIowait
	event.Metrics["percent_steal"] = ts.PercentSteal
	event.Metrics["percent_idle"] = ts.PercentIdle

	return event
}

type MemoryTimeslice struct {
	TimestampMs   int64   `json:"timestamp_ms"`
	MemFreeK      int64   `json:"free_k"`
	AvailK        int64   `json:"avail_k"`
	UsedK         int64   `json:"used_k"`
	UsedPercent   float64 `json:"used_percent"`
	BuffersK      int64   `json:"buffers_k"`
	CachedK       int64   `json:"cached_k"`
	CommitK       int64   `json:"commit_k"`
	CommitPercent float64 `json:"commit_percent"`
	ActiveK       int64   `json:"active_k"`
	InactiveK     int64   `json:"inactive_k"`
	DirtyK        int64   `json:"dirty_k"`
}

func (ts *MemoryTimeslice) toMetricsEvent() *MetricsEvent {
	event := &MetricsEvent{
		Timestamp: timeutil.MillisecondsToTime(ts.TimestampMs),
		Metrics:   MetricSet{},
	}

	event.Metrics["free_k"] = ts.MemFreeK
	event.Metrics["avail_k"] = ts.AvailK
	event.Metrics["used_k"] = ts.UsedK
	event.Metrics["used_percent"] = ts.UsedPercent
	event.Metrics["buffers_k"] = ts.BuffersK
	event.Metrics["cached_k"] = ts.CachedK
	event.Metrics["commit_k"] = ts.CommitK
	event.Metrics["commit_percent"] = ts.CommitPercent
	event.Metrics["active_k"] = ts.ActiveK
	event.Metrics["inactive_k"] = ts.InactiveK
	event.Metrics["dirty_k"] = ts.DirtyK

	return event
}

type ProcessTimeslice struct {
	TimestampMs     int64   `json:"timestamp_ms"`
	RunQueueSize    int64   `json:"run_queue_size"`
	ProcessListSize int64   `json:"process_list_size"`
	LoadAverage1m   float64 `json:"load_average_1m"`
	LoadAverage5m   float64 `json:"load_average_5m`
	LoadAverage15m  float64 `json:"load_average_15m"`
	Blocked         int64   `json:"blocked"`
}

func (ts *ProcessTimeslice) toMetricsEvent() *MetricsEvent {
	event := &MetricsEvent{
		Timestamp: timeutil.MillisecondsToTime(ts.TimestampMs),
		Metrics:   MetricSet{},
	}

	event.Metrics["run_queue_size"] = ts.RunQueueSize
	event.Metrics["process_list_size"] = ts.ProcessListSize
	event.Metrics["load_average_1m"] = ts.LoadAverage1m
	event.Metrics["load_average_5m"] = ts.LoadAverage5m
	event.Metrics["load_average_15m"] = ts.LoadAverage15m
	event.Metrics["blocked"] = ts.Blocked

	return event
}

type ZitiFabricMeshSummary struct {
	TimestampMs int64                   `json:"timestamp_ms"`
	RouterIds   []string                `json:"router_ids"`
	Links       []ZitiFabricLinkSummary `json:"links,omitempty"`
}

type ZitiFabricLinkSummary struct {
	LinkId      string  `json:"link_id"`
	State       string  `json:"state"`
	SrcRouterId string  `json:"src_router_id"`
	SrcLatency  float64 `json:"src_latency"`
	DstRouterId string  `json:"dst_router_id"`
	DstLatency  float64 `json:"dst_latency"`
}

type ZitiFabricRouterMetricsSummary struct {
	SourceId             string                         `json:"source_id"`
	TimestampMs          int64                          `json:"timestamp_ms"`
	FabricRxBytesRateM1  float64                        `json:"fabric_rx_bytes_rate_m1"`
	FabricTxBytesRateM1  float64                        `json:"fabric_tx_bytes_rate_m1"`
	IngressRxBytesRateM1 float64                        `json:"ingress_rx_bytes_rate_m1"`
	IngressTxBytesRateM1 float64                        `json:"ingress_tx_bytes_rate_m1"`
	EgressRxBytesRateM1  float64                        `json:"egress_rx_bytes_rate_m1"`
	EgressTxBytesRateM1  float64                        `json:"egress_tx_bytes_rate_m1"`
	Links                []ZitiFabricLinkMetricsSummary `json:"links,omitempty"`
}

type ZitiFabricLinkMetricsSummary struct {
	LinkId        string  `json:"link_id"`
	LatencyP9999  float64 `json:"latency_p9999"`
	RxBytesRateM1 float64 `json:"rx_bytes_rate_m1"`
	TxBytesRateM1 float64 `json:"tx_bytes_rate_m1"`
}

type IperfSummary struct {
	Timeslices    []*IperfTimeslice `json:"timeslices"`
	Bytes         float64           `json:"bytes"`
	BitsPerSecond float64           `json:"bits_per_second"`
}

type IperfTimeslice struct {
	TimestampMs   int64   `json:"timestamp_ms"`
	BitsPerSecond float64 `json:"bits_per_second"`
}

type IperfUdpSummary struct {
	Timeslices    []*IperfUdpTimeslice `json:"timeslices"`
	Bytes         float64              `json:"bytes"`
	BitsPerSecond float64              `json:"bits_per_second"`
	JitterMs      float64              `json:"jitter_ms"`
	LostPackets   float64              `json:"lost_packets"`
}

type IperfUdpTimeslice struct {
	TimestampMs   int64   `json:"timestamp_ms"`
	BitsPerSecond float64 `json:"bits_per_second"`
	Packets       float64 `json:"packets"`
}
