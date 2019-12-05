package console

import (
	"github.com/netfoundry/ziti-foundation/channel2"
	"github.com/netfoundry/ziti-fabric/pb/mgmt_pb"
	"fmt"
	"github.com/golang/protobuf/proto"
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
		logrus.Fatalf("error handling metrics receive (%w)", err)
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
