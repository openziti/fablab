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
	"github.com/golang/protobuf/proto"
	"github.com/openziti/fabric/pb/mgmt_pb"
	"github.com/openziti/foundation/channel2"
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
		logrus.Fatalf("error handling receive routers list (%v)", err)
	}
	mgmtRouters.server.Routers(response.Routers)
	logrus.Infof("updated [%d] routers", len(response.Routers))
}

type mgmtRouters struct {
	server *Server
}
