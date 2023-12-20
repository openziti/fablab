/*
	Copyright NetFoundry Inc.

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

package operation

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/foundation/v2/errorz"
	log "github.com/sirupsen/logrus"
	"net/url"
	"time"
)

func Influx2MetricsReporter() model.Stage {
	return &influx2MetricsReporterStage{}
}

type influx2MetricsReporterStage struct {
	errorz.ErrorHolderImpl
}

func (stage *influx2MetricsReporterStage) Execute(run model.Run) error {
	m := run.GetModel()
	urlVar := m.GetRequiredStringVariable(stage, "metrics.influxdb.url")
	db := m.GetRequiredStringVariable(stage, "metrics.influxdb.db")
	token := m.GetRequiredStringVariable(stage, "metrics.influxdb.token")
	org := m.GetRequiredStringVariable(stage, "metrics.influxdb.org")
	bucket := m.GetRequiredStringVariable(stage, "metrics.influxdb.bucket")

	if stage.HasError() {
		return stage.GetError()
	}

	influxUrl, err := url.Parse(urlVar)
	if err != nil {
		return err
	}

	config := influx2Config{
		url:      *influxUrl,
		database: db,
		token:    token,
		org:      org,
		bucket:   bucket,
	}

	handler, err := NewInflux2DBMetricsHandler(&config)
	if err != nil {
		return err
	}

	m.MetricsHandlers = append(m.MetricsHandlers, handler)
	return nil
}

func NewInflux2DBMetricsHandler(cfg *influx2Config) (model.MetricsHandler, error) {
	rep := &influx2Reporter{
		url:         cfg.url,
		database:    cfg.bucket,
		token:       cfg.token,
		metricsChan: make(chan *hostMetricsEvent2, 10),
		org:         cfg.org,
	}

	if err := rep.makeClient(); err != nil {
		return nil, fmt.Errorf("unable to make HandlerTypeInfluxDB influxdb. err=%v", err)
	}

	go rep.run()
	return rep, nil
}

type influx2Reporter struct {
	url         url.URL
	database    string
	metricsChan chan *hostMetricsEvent2
	client      influxdb2.Client
	token       string
	org         string
}

func (reporter *influx2Reporter) AcceptHostMetrics(host *model.Host, event *model.MetricsEvent) {
	reporter.metricsChan <- &hostMetricsEvent2{
		host:  host,
		event: event,
	}
}

func (reporter *influx2Reporter) makeClient() error {
	reporter.client = influxdb2.NewClient(reporter.url.String(), reporter.token)
	return nil
}

func (reporter *influx2Reporter) run() {
	logz := pfxlog.Logger()
	logz.Info("started")
	defer logz.Warn("exited")

	pingTicker := time.NewTicker(time.Second * 5)
	defer pingTicker.Stop()

	for {
		select {
		case msg := <-reporter.metricsChan:
			if err := reporter.send(msg); err != nil {
				logz.Printf("unable to send metrics to influxdb. err=%v", err)
			}
		case <-pingTicker.C:
			_, err := reporter.client.Ping(context.Background())
			if err != nil {
				logz.Printf("got error while sending a ping to influxdb, trying to recreate influxdb. err=%v", err)

				if err = reporter.makeClient(); err != nil {
					logz.Printf("unable to make client connection to influxdb. err=%v", err)
				}
			}
		}
	}
}

func AsBatch2(hostEvent *hostMetricsEvent2) ([]*write.Point, error) {
	var pts []*write.Point

	event := hostEvent.event

	tags := make(map[string]string)

	tags["source"] = hostEvent.host.GetId()
	tags["publicIp"] = hostEvent.host.PublicIp

	for k, v := range hostEvent.event.Tags {
		tags[k] = v
	}

	event.Metrics.VisitUngroupedMetrics(func(name string, val interface{}) {
		p := influxdb2.NewPoint(name, tags, map[string]interface{}{"value": val}, event.Timestamp)
		pts = append(pts, p)
	})

	event.Metrics.VisitGroupedMetrics(func(name string, group model.MetricSet) {
		p := influxdb2.NewPoint(name, tags, group, event.Timestamp)
		pts = append(pts, p)
	})

	return pts, nil
}

func (reporter *influx2Reporter) send(msg *hostMetricsEvent2) error {
	points, err := AsBatch2(msg)
	if err != nil {
		return err
	}
	log.Printf("org: %s, bucket: %s", reporter.org, reporter.database)
	writeAPI := reporter.client.WriteAPI(reporter.org, reporter.database) //reporter.database = bucket name
	for _, p := range points {
		writeAPI.WritePoint(p)
	}

	writeAPI.Flush()
	return nil
}

type influx2Config struct {
	url      url.URL
	database string
	token    string
	org      string
	bucket   string
}

type hostMetricsEvent2 struct {
	host  *model.Host
	event *model.MetricsEvent
}
