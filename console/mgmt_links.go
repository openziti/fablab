package console

import (
	"github.com/netfoundry/ziti-foundation/channel2"
	"github.com/netfoundry/ziti-fabric/pb/mgmt_pb"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

func newMgmtLinks(server *Server) *mgmtLinks {
	return &mgmtLinks{server: server}
}

func (mgmtLinks *mgmtLinks) ContentType() int32 {
	return int32(mgmt_pb.ContentType_ListLinksResponseType)
}

func (mgmtLinks *mgmtLinks) HandleReceive(msg *channel2.Message, ch channel2.Channel) {
	response := &mgmt_pb.ListLinksResponse{}
	err := proto.Unmarshal(msg.Body, response)
	if err != nil {
		logrus.Fatalf("error handling receive links list (%w)", err)
	}
	mgmtLinks.server.Links(response.Links)
	logrus.Infof("updated [%d] links", len(response.Links))
}

type mgmtLinks struct {
	server *Server
}