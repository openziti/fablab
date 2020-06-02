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
		logrus.Fatalf("error handling receive links list (%v)", err)
	}
	mgmtLinks.server.Links(response.Links)
	logrus.Infof("updated [%d] links", len(response.Links))
}

type mgmtLinks struct {
	server *Server
}
