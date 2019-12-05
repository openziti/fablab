package metrics

import (
	"github.com/netfoundry/fablab/kernel"
	"github.com/netfoundry/ziti-foundation/channel2"
	"github.com/netfoundry/ziti-fabric/pb/mgmt_pb"
	"github.com/netfoundry/ziti-foundation/identity/dotziti"
	"github.com/netfoundry/ziti-foundation/transport"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/michaelquigley/pfxlog"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

func Metrics() kernel.Action {
	return &metrics{}
}

func (metrics *metrics) Execute(m *kernel.Model) error {
	var mgmt channel2.Channel
	if endpoint, id, err := dotziti.LoadIdentity("fablab"); err == nil {
		if address, err := transport.ParseAddress(endpoint); err == nil {
			dialer := channel2.NewClassicDialer(id, address, nil)
			if ch, err := channel2.NewChannel("mgmt", dialer, nil); err == nil {
				mgmt = ch
			} else {
				return fmt.Errorf("error connecting mgmt channel (%w)", err)
			}
		} else {
			return fmt.Errorf("invalid endpoint address (%w)", err)
		}
	} else {
		return fmt.Errorf("unable to load 'fablab' identity (%w)", err)
	}

	mgmt.AddReceiveHandler(metrics)

	request := &mgmt_pb.StreamMetricsRequest{
		Matchers: []*mgmt_pb.StreamMetricsRequest_MetricMatcher{
			&mgmt_pb.StreamMetricsRequest_MetricMatcher{},
		},
	}
	body, err := proto.Marshal(request)
	if err != nil {
		logrus.Fatalf("error marshaling metrics request (%w)", err)
	}

	requestMsg := channel2.NewMessage(int32(mgmt_pb.ContentType_StreamMetricsRequestType), body)
	waitCh, err := mgmt.SendAndSync(requestMsg)
	if err != nil {
		logrus.Fatalf("error sending metrics request (%w)", err)
	}
	select {
	case err := <-waitCh:
		if err != nil {
			logrus.Fatalf("error waiting for response (%w)", err)
		}
	case <-time.After(5 * time.Second):
		logrus.Fatal("timeout")
	}

	waitForChannelClose(mgmt)

	return nil
}

func (metrics *metrics) ContentType() int32 {
	return int32(mgmt_pb.ContentType_StreamMetricsEventType)
}

func (metrics *metrics) HandleReceive(msg *channel2.Message, ch channel2.Channel) {
	response := &mgmt_pb.StreamMetricsEvent{}
	err := 	proto.Unmarshal(msg.Body, response)
	if err != nil {
		panic(err)
	}

	/*
	fmt.Printf("%v - source(%v)\n", formattedTimestamp(response.Timestamp), response.SourceId)
	fmt.Printf("\tTags: %v\n", response.Tags)

	var keys []string
	var outputMap = make(map[string]string)
	for name, value := range response.IntMetrics {
		outputMap[name] = fmt.Sprintf("%v=%v", name, value)
		keys = append(keys, name)
	}

	for name, value := range response.FloatMetrics {
		outputMap[name] = fmt.Sprintf("%v=%v", name, value)
		keys = append(keys, name)
	}

	sort.Strings(keys)

	for _, key := range keys {
		fmt.Println(outputMap[key])
	}

	for _, bucket := range response.IntervalMetrics {
		fmt.Printf("%v: (%v) -> (%v)\n", bucket.Name, formattedTimestamp(bucket.IntervalStartUTC), formattedTimestamp(bucket.IntervalEndUTC))
		for name, value := range bucket.Values {
			fmt.Printf("\t%v=%v\n", name, value)
		}
	}
	*/

	logrus.Infof("source = [%s]", response.SourceId)

	fmt.Println()
}

func formattedTimestamp(protobufTS *timestamp.Timestamp) string {
	ts, err := ptypes.Timestamp(protobufTS)
	if err != nil {
		panic(err)
	}
	return ts.Format(time.RFC3339)
}

type metrics struct {
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
	pfxlog.Logger().Info("Management channel to controller closed. Shutting down.")
	watcher.waitGroup.Done()
}
