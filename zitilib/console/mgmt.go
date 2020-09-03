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
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/fabric/pb/mgmt_pb"
	"github.com/openziti/foundation/channel2"
	"github.com/openziti/foundation/identity/dotziti"
	"github.com/openziti/foundation/transport"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

func newMgmt(server *Server) *mgmt {
	return &mgmt{
		server: server,
	}
}

func (mgmt *mgmt) execute() error {
	if endpoint, id, err := dotziti.LoadIdentity(model.ActiveInstanceId()); err == nil {
		if address, err := transport.ParseAddress(endpoint); err == nil {
			dialer := channel2.NewClassicDialer(id, address, nil)
			if ch, err := channel2.NewChannel("mgmt", dialer, nil); err == nil {
				mgmt.ch = ch
			} else {
				return fmt.Errorf("error connecting mgmt channel (%w)", err)
			}
		} else {
			return fmt.Errorf("invalid endpoint address (%w)", err)
		}
	} else {
		return fmt.Errorf("unable to load 'fablab' identity (%w)", err)
	}

	mgmt.ch.AddReceiveHandler(newMgmtMetrics(mgmt.server))
	mgmt.ch.AddReceiveHandler(newMgmtRouters(mgmt.server))
	mgmt.ch.AddReceiveHandler(newMgmtLinks(mgmt.server))
	go mgmt.pollNetworkShape()

	request := &mgmt_pb.StreamMetricsRequest{
		Matchers: []*mgmt_pb.StreamMetricsRequest_MetricMatcher{
			&mgmt_pb.StreamMetricsRequest_MetricMatcher{},
		},
	}
	body, err := proto.Marshal(request)
	if err != nil {
		logrus.Fatalf("error marshaling metrics request (%v)", err)
	}

	requestMsg := channel2.NewMessage(int32(mgmt_pb.ContentType_StreamMetricsRequestType), body)
	errCh, err := mgmt.ch.SendAndSync(requestMsg)
	if err != nil {
		logrus.Fatalf("error queuing metrics request (%v)", err)
	}
	select {
	case err := <-errCh:
		if err != nil {
			logrus.Fatalf("error sending metrics request (%v)", err)
		}
	case <-time.After(5 * time.Second):
		logrus.Fatal("timeout")
	}

	waitForChannelClose(mgmt.ch)

	return nil
}

func (mgmt *mgmt) pollNetworkShape() {
	for {
		routersRequest := &mgmt_pb.ListRoutersRequest{}
		body, err := proto.Marshal(routersRequest)
		if err != nil {
			logrus.Fatalf("error marshaling list routers request (%v)", err)
		}
		routersRequestMsg := channel2.NewMessage(int32(mgmt_pb.ContentType_ListRoutersRequestType), body)
		err = mgmt.ch.Send(routersRequestMsg)
		if err != nil {
			logrus.Fatalf("error queuing list routers request (%v)", err)
		}

		linksRequest := &mgmt_pb.ListLinksRequest{}
		body, err = proto.Marshal(linksRequest)
		if err != nil {
			logrus.Fatalf("error marshaling list links request (%v)", err)
		}
		linksRequestMsg := channel2.NewMessage(int32(mgmt_pb.ContentType_ListLinksRequestType), body)
		err = mgmt.ch.Send(linksRequestMsg)
		if err != nil {
			logrus.Fatalf("error queuing list links request (%v)", err)
		}

		time.Sleep(5 * time.Second)
	}
}

func waitForChannelClose(ch channel2.Channel) {
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)

	ch.AddCloseHandler(&closeWatcher{waitGroup})

	waitGroup.Wait()
}

type closeWatcher struct {
	waitGroup *sync.WaitGroup
}

func (watcher *closeWatcher) HandleClose(ch channel2.Channel) {
	watcher.waitGroup.Done()
}

type mgmt struct {
	ch     channel2.Channel
	server *Server
}
