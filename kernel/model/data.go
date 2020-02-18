package model

type HostSummary struct {
	Cpu     []*CpuTimeslice     `json:"cpu,omitempty"`
	Memory  []*MemoryTimeslice  `json:"memory,omitempty"`
	Process []*ProcessTimeslice `json:"process,omitempty"`
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

type ProcessTimeslice struct {
	TimestampMs     int64   `json:"timestamp_ms"`
	RunQueueSize    int64   `json:"run_queue_size"`
	ProcessListSize int64   `json:"process_list_size"`
	LoadAverage1m   float64 `json:"load_average_1m"`
	LoadAverage5m   float64 `json:"load_average_5m`
	LoadAverage15m  float64 `json:"load_average_15m"`
	Blocked         int64   `json:"blocked"`
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
