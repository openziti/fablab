package model

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
	TimestampMs   int64   `json:"timestamp_ms"`
	BitsPerSecond float64 `json:"bits_per_second"`
}
