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

package zitilib_runlevel_5_operation

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fabric/pb/mgmt_pb"
	"github.com/openziti/foundation/channel2"
	"github.com/openziti/foundation/identity/dotziti"
	"github.com/openziti/foundation/transport"
	"github.com/sirupsen/logrus"
	"time"
)

func ModelMetrics(closer chan struct{}) model.OperatingStage {
	return MetricsWithIdMapper(closer, func(id string) string {
		return "#" + id
	})
}

func ModelMetricsWithIdMapper(closer <-chan struct{}, f func(string) string) model.OperatingStage {
	return &modelMetrics{
		closer:             closer,
		idToSelectorMapper: f,
	}
}

type modelMetrics struct {
	ch                 channel2.Channel
	m                  *model.Model
	closer             <-chan struct{}
	idToSelectorMapper func(string) string
}

func (metrics *modelMetrics) Operate(run model.Run) error {
	if endpoint, id, err := dotziti.LoadIdentity(model.ActiveInstanceId()); err == nil {
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
		Matchers: []*mgmt_pb.StreamMetricsRequest_MetricMatcher{},
	}
	body, err := proto.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshaling metrics request (%w)", err)
	}

	requestMsg := channel2.NewMessage(int32(mgmt_pb.ContentType_StreamMetricsRequestType), body)
	err = metrics.ch.SendWithTimeout(requestMsg, 5*time.Second)
	if err != nil {
		logrus.Fatalf("error queuing metrics request (%v)", err)
	}

	metrics.m = run.GetModel()
	go metrics.runMetrics()

	return nil
}

func (metrics *modelMetrics) ContentType() int32 {
	return int32(mgmt_pb.ContentType_StreamMetricsEventType)
}

func (metrics *modelMetrics) HandleReceive(msg *channel2.Message, _ channel2.Channel) {
	response := &mgmt_pb.StreamMetricsEvent{}
	err := proto.Unmarshal(msg.Body, response)
	if err != nil {
		logrus.Error("error handling metrics receive (%w)", err)
	}

	hostSelector := metrics.idToSelectorMapper(response.SourceId)
	host, err := metrics.m.SelectHost(hostSelector)
	if err == nil {
		modelEvent := metrics.toModelMetricsEvent(response)
		metrics.m.AcceptHostMetrics(host, modelEvent)
		logrus.Infof("<$= [%s]", response.SourceId)
	} else {
		logrus.Errorf("unable to find host (%v)", err)
	}
}

func (metrics *modelMetrics) runMetrics() {
	logrus.Infof("starting")
	defer logrus.Infof("exiting")

	<-metrics.closer
	_ = metrics.ch.Close()
}

func (metrics *modelMetrics) toModelMetricsEvent(fabricEvent *mgmt_pb.StreamMetricsEvent) *model.MetricsEvent {
	modelEvent := &model.MetricsEvent{
		Timestamp: time.Unix(fabricEvent.Timestamp.Seconds, int64(fabricEvent.Timestamp.Nanos)),
		Metrics:   model.MetricSet{},
	}

	for name, val := range fabricEvent.IntMetrics {
		group := fabricEvent.MetricGroup[name]
		modelEvent.Metrics.AddGroupedMetric(group, name, val)
	}

	for name, val := range fabricEvent.FloatMetrics {
		group := fabricEvent.MetricGroup[name]
		modelEvent.Metrics.AddGroupedMetric(group, name, val)
	}

	return modelEvent
}
