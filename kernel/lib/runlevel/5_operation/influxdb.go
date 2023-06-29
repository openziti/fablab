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
	"errors"
	"fmt"
	influxdb "github.com/influxdata/influxdb1-client"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/foundation/v2/errorz"
	"net/url"
	"time"
)

func InfluxMetricsReporter() model.Stage {
	return &influxMetricsReporterStage{}
}

type influxMetricsReporterStage struct {
	errorz.ErrorHolderImpl
}

func (stage *influxMetricsReporterStage) Execute(run model.Run) error {
	m := run.GetModel()
	urlVar := m.GetRequiredStringVariable(stage, "metrics.influxdb.url")
	db := m.GetRequiredStringVariable(stage, "metrics.influxdb.db")
	username := m.GetRequiredStringVariable(stage, "credentials.influxdb.username")
	password := m.GetRequiredStringVariable(stage, "credentials.influxdb.password")

	if stage.HasError() {
		return stage.GetError()
	}

	influxUrl, err := url.Parse(urlVar)
	if err != nil {
		return err
	}

	config := influxConfig{
		url:      *influxUrl,
		database: db,
		username: username,
		password: password,
	}

	handler, err := NewInfluxDBMetricsHandler(&config)
	if err != nil {
		return err
	}

	m.MetricsHandlers = append(m.MetricsHandlers, handler)
	return nil
}

func NewInfluxDBMetricsHandler(cfg *influxConfig) (model.MetricsHandler, error) {
	rep := &influxReporter{
		url:         cfg.url,
		database:    cfg.database,
		username:    cfg.username,
		password:    cfg.password,
		metricsChan: make(chan *hostMetricsEvent, 10),
	}

	if err := rep.makeClient(); err != nil {
		return nil, fmt.Errorf("unable to make HandlerTypeInfluxDB influxdb. err=%v", err)
	}

	go rep.run()
	return rep, nil
}

type influxReporter struct {
	url         url.URL
	database    string
	username    string
	password    string
	metricsChan chan *hostMetricsEvent

	client *influxdb.Client
}

func (reporter *influxReporter) AcceptHostMetrics(host *model.Host, event *model.MetricsEvent) {
	reporter.metricsChan <- &hostMetricsEvent{
		host:  host,
		event: event,
	}
}

func (reporter *influxReporter) makeClient() (err error) {
	reporter.client, err = influxdb.NewClient(influxdb.Config{
		URL:      reporter.url,
		Username: reporter.username,
		Password: reporter.password,
	})

	return
}

func (reporter *influxReporter) run() {
	log := pfxlog.Logger()
	log.Info("started")
	defer log.Warn("exited")

	pingTicker := time.NewTicker(time.Second * 5)
	defer pingTicker.Stop()

	for {
		select {
		case msg := <-reporter.metricsChan:
			if err := reporter.send(msg); err != nil {
				log.Printf("unable to send metrics to influxdb. err=%v", err)
			}
		case <-pingTicker.C:
			_, _, err := reporter.client.Ping()
			if err != nil {
				log.Printf("got error while sending a ping to influxdb, trying to recreate influxdb. err=%v", err)

				if err = reporter.makeClient(); err != nil {
					log.Printf("unable to make client connection to influxdb. err=%v", err)
				}
			}
		}
	}
}

func AsBatch(hostEvent *hostMetricsEvent) (*influxdb.BatchPoints, error) {
	var pts []influxdb.Point

	event := hostEvent.event

	tags := make(map[string]string)

	tags["source"] = hostEvent.host.GetId()
	tags["publicIp"] = hostEvent.host.PublicIp

	for k, v := range hostEvent.event.Tags {
		tags[k] = v
	}

	event.Metrics.VisitUngroupedMetrics(func(name string, val interface{}) {
		pts = append(pts, influxdb.Point{
			Measurement: name,
			Tags:        tags,
			Fields: map[string]interface{}{
				"value": val,
			},
			Time: event.Timestamp,
		})
	})

	event.Metrics.VisitGroupedMetrics(func(name string, group model.MetricSet) {
		pts = append(pts, influxdb.Point{
			Measurement: name,
			Tags:        tags,
			Fields:      group,
			Time:        event.Timestamp,
		})
	})

	bps := &influxdb.BatchPoints{
		Points: pts,
	}

	return bps, nil
}

func (reporter *influxReporter) send(msg *hostMetricsEvent) error {
	bps, err := AsBatch(msg)
	if err != nil {
		return err
	}

	bps.Database = reporter.database
	_, err = reporter.client.Write(*bps)
	return err
}

type influxConfig struct {
	url      url.URL
	database string
	username string
	password string
}

func LoadInfluxConfig(src map[interface{}]interface{}) (*influxConfig, error) {
	cfg := &influxConfig{}

	if value, found := src["url"]; found {
		if urlSrc, ok := value.(string); ok {
			if url, err := url.Parse(urlSrc); err == nil {
				cfg.url = *url
			} else {
				return nil, fmt.Errorf("cannot parse influx 'url' value (%s)", err)
			}
		} else {
			return nil, errors.New("invalid influx 'url' value")
		}
	} else {
		return nil, errors.New("missing influx 'url' config")
	}

	if value, found := src["database"]; found {
		if database, ok := value.(string); ok {
			cfg.database = database
		} else {
			return nil, errors.New("invalid influx 'database' value")
		}
	} else {
		return nil, errors.New("missing influx 'database' config")
	}

	if value, found := src["username"]; found {
		if username, ok := value.(string); ok {
			cfg.username = username
		} else {
			return nil, errors.New("invalid influx 'username' value")
		}
	}

	if value, found := src["password"]; found {
		if password, ok := value.(string); ok {
			cfg.password = password
		} else {
			return nil, errors.New("invalid influx 'password' value")
		}
	}

	return cfg, nil
}

type hostMetricsEvent struct {
	host  *model.Host
	event *model.MetricsEvent
}
