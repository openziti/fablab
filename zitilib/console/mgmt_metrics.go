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

package console

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/openziti/fabric/pb/mgmt_pb"
	"github.com/openziti/foundation/channel2"
	"github.com/sirupsen/logrus"
	"strings"
)

func newMgmtMetrics(server *Server) *mgmtMetrics {
	return &mgmtMetrics{server: server}
}

func (mgmt *mgmtMetrics) ContentType() int32 {
	return int32(mgmt_pb.ContentType_StreamMetricsEventType)
}

func (mgmt *mgmtMetrics) HandleReceive(msg *channel2.Message, ch channel2.Channel) {
	response := &mgmt_pb.StreamMetricsEvent{}
	err := proto.Unmarshal(msg.Body, response)
	if err != nil {
		logrus.Fatalf("error handling metrics receive (%v)", err)
	}

	wsMsg := &Message{Source: response.SourceId}
	for k, v := range response.FloatMetrics {
		if strings.Index(k, "bytesrate.mean_rate") != -1 {
			wsMsg.Metrics = append(wsMsg.Metrics, &Metric{Key: k, Value: fmt.Sprintf("%0.2f", v)})
		}
	}

	mgmt.server.SendAll(wsMsg)
	logrus.Infof("sent [%s=[%s]]", "source", response.SourceId)
}

type mgmtMetrics struct {
	server *Server
}
