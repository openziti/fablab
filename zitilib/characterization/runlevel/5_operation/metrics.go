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

package __operation

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/ziti-fabric/pb/mgmt_pb"
	"github.com/netfoundry/ziti-foundation/channel2"
	"github.com/netfoundry/ziti-foundation/identity/dotziti"
	"github.com/netfoundry/ziti-foundation/transport"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

func Metrics(closer chan struct{}) model.OperatingStage {
	return &metrics{closer: closer}
}

func (metrics *metrics) Operate(m *model.Model) error {
	if endpoint, id, err := dotziti.LoadIdentity("fablab"); err == nil {
		if address, err := transport.ParseAddress(endpoint); err == nil {
			dialer := channel2.NewClassicDialer(id, address, nil)
			if ch, err := channel2.NewChannel("metrics", dialer, nil); err == nil {
				metrics.ch = ch
			} else {
				return fmt.Errorf("error connecting metrics channel (%w)", err)
			}
		} else {
			return fmt.Errorf("invalid endpoint address (%w)", err)
		}
	} else {
		return fmt.Errorf("unable to load 'fablab' identity (%w)", err)
	}

	metrics.ch.AddReceiveHandler(metrics)

	request := &mgmt_pb.StreamMetricsRequest{
		Matchers: []*mgmt_pb.StreamMetricsRequest_MetricMatcher{
			&mgmt_pb.StreamMetricsRequest_MetricMatcher{},
		},
	}
	body, err := proto.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshaling metrics request (%w)", err)
	}

	requestMsg := channel2.NewMessage(int32(mgmt_pb.ContentType_StreamMetricsRequestType), body)
	errCh, err := metrics.ch.SendAndSync(requestMsg)
	if err != nil {
		logrus.Fatalf("error queuing metrics request (%w)", err)
	}
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error sending metrics request (%w)", err)
		}

	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout")
	}

	metrics.m = m
	go metrics.runMetrics()

	return nil
}

func (metrics *metrics) ContentType() int32 {
	return int32(mgmt_pb.ContentType_StreamMetricsEventType)
}

func (metrics *metrics) HandleReceive(msg *channel2.Message, _ channel2.Channel) {
	response := &mgmt_pb.StreamMetricsEvent{}
	err := proto.Unmarshal(msg.Body, response)
	if err != nil {
		logrus.Error("error handling metrics receive (%w)", err)
	}

	host, err := metrics.m.GetHostById(response.SourceId)
	if err == nil {
		if host.Data == nil {
			host.Data = make(map[string]interface{})
		}
		if _, found := host.Data["fabric_metrics"]; !found {
			host.Data["fabric_metrics"] = make([]model.ZitiFabricRouterMetricsSummary, 0)
		}

		summary, err := SummarizeZitiFabricMetrics(response)
		if err == nil {
			summaries := host.Data["fabric_metrics"].([]model.ZitiFabricRouterMetricsSummary)
			summaries = append(summaries, summary)
			host.Data["fabric_metrics"] = summaries

			logrus.Infof("<$= [%s]", response.SourceId)
		}

	} else {
		logrus.Errorf("unable to find host (%w)", err)
	}
}

func (metrics *metrics) runMetrics() {
	logrus.Infof("starting")
	defer logrus.Infof("exiting")

	for {
		select {
		case <-metrics.closer:
			_ = metrics.ch.Close()
			return
		}
	}
}

type metrics struct {
	ch     channel2.Channel
	m      *model.Model
	closer chan struct{}
}

func SummarizeZitiFabricMetrics(metrics *mgmt_pb.StreamMetricsEvent) (model.ZitiFabricRouterMetricsSummary, error) {
	summary := model.ZitiFabricRouterMetricsSummary{
		SourceId:    metrics.SourceId,
		TimestampMs: fablib.TimeToMilliseconds(time.Unix(metrics.Timestamp.Seconds, int64(metrics.Timestamp.Nanos))),
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