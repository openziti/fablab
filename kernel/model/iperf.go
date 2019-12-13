package model

type IperfSummary struct {
	TimestampMs   int64   `json:"timestamp_ms"`
	BitsPerSecond float64 `json:"bits_per_second"`
}
