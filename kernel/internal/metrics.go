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
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/ziti-fabric/pb/mgmt_pb"
	"strings"
	"time"
)

func SummarizeZitiFabricMetrics(metrics *mgmt_pb.StreamMetricsEvent) (model.ZitiFabricRouterMetricsSummary, error) {
	summary := model.ZitiFabricRouterMetricsSummary{
		SourceId:    metrics.SourceId,
		TimestampMs: TimeToMilliseconds(time.Unix(metrics.Timestamp.Seconds, int64(metrics.Timestamp.Nanos))),
	}

	if value, found := metrics.FloatMetrics["fabric.rx.bytesrate.m1_rate"]; found {
		summary.FabricRxBytesRateM1 = value
	}
	if value, found := metrics.FloatMetrics["fabric.tx.bytesrate.m1_rate"]; found {
		summary.FabricTxBytesRateM1 = value
	}
	if value, found := metrics.FloatMetrics["ingress.rx.bytesrate.m1_rate"]; found {
		summary.IngressRxBytesRateM1 = value
	}
	if value, found := metrics.FloatMetrics["ingress.tx.bytesrate.m1_rate"]; found {
		summary.IngressTxBytesRateM1 = value
	}
	if value, found := metrics.FloatMetrics["egress.rx.bytesrate.m1_rate"]; found {
		summary.EgressRxBytesRateM1 = value
	}
	if value, found := metrics.FloatMetrics["egress.tx.bytesrate.m1_rate"]; found {
		summary.EgressTxBytesRateM1 = value
	}

	for _, linkId := range linkIdsFromMetrics(metrics) {
		linkSummary := model.ZitiFabricLinkMetricsSummary{LinkId: linkId}

		if value, found := metrics.FloatMetrics["link."+linkId+".latency.p9999"]; found {
			linkSummary.LatencyP9999 = value
		}
		if value, found := metrics.FloatMetrics["link."+linkId+".rx.bytesrate.m1_rate"]; found {
			linkSummary.RxBytesRateM1 = value
		}
		if value, found := metrics.FloatMetrics["link."+linkId+".tx.bytesrate.m1_rate"]; found {
			linkSummary.TxBytesRateM1 = value
		}

		summary.Links = append(summary.Links, linkSummary)
	}

	return summary, nil
}

func linkIdsFromMetrics(metrics *mgmt_pb.StreamMetricsEvent) []string {
	visitedLinks := make(map[string]struct{})
	for k := range metrics.FloatMetrics {
		if strings.HasPrefix(k, "link.") {
			linkId := linkIdFromMetricKey(k)
			visitedLinks[linkId] = struct{}{}
		}
	}

	linkIds := make([]string, 0)
	for k := range visitedLinks {
		linkIds = append(linkIds, k)
	}

	return linkIds
}

func linkIdFromMetricKey(metricKey string) string {
	linkId := metricKey[5:]
	endIdx := strings.IndexRune(linkId, '.')
	if endIdx > -1 {
		linkId = linkId[:endIdx]
	}
	return linkId
}
