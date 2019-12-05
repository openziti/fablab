package console

import (
	"github.com/golang/protobuf/proto"
	"github.com/netfoundry/ziti-fabric/pb/mgmt_pb"
	"github.com/netfoundry/ziti-foundation/channel2"
	"github.com/sirupsen/logrus"
)

func newMgmtRouters(server *Server) *mgmtRouters {
	return &mgmtRouters{server: server}
}

func (mgmtRouters *mgmtRouters) ContentType() int32 {
	return int32(mgmt_pb.ContentType_ListRoutersResponseType)
}

func (mgmtRouters *mgmtRouters) HandleReceive(msg *channel2.Message, ch channel2.Channel) {
	response := &mgmt_pb.ListRoutersResponse{}
	err := proto.Unmarshal(msg.Body, response)
	if err != nil {
		logrus.Fatalf("error handling receive routers list (%w)", err)
	}
	mgmtRouters.server.Routers(response.Routers)
	logrus.Infof("updated [%d] routers", len(response.Routers))
}

type mgmtRouters struct {
	server *Server
}