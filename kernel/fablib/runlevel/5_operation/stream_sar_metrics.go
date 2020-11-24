/*
	Copyright 2020 NetFoundry, Inc.

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
	"fmt"
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/foundation/util/concurrenz"
	"github.com/sirupsen/logrus"
)

func StreamSarMetrics(host *model.Host, intervalSeconds, reportIntervalCount int, runPhase Phase, cleanupPhase Phase) model.OperatingStage {
	return &streamSarMetrics{
		host:                host,
		intervalSeconds:     intervalSeconds,
		reportIntervalCount: reportIntervalCount,
		closer:              runPhase.GetCloser(),
		joiner:              cleanupPhase.AddJoiner(),
	}
}

type streamSarMetrics struct {
	host                *model.Host
	intervalSeconds     int
	reportIntervalCount int
	joiner              chan struct{}
	closer              <-chan struct{}
	closed              concurrenz.AtomicBoolean
}

func (s *streamSarMetrics) Operate(run model.Run) error {
	m := run.GetModel()
	go s.waitForClose(run)
	ssh := fablib.NewSshConfigFactoryImpl(m, s.host.PublicIp)
	go s.runSar(ssh)
	return nil
}

func (s *streamSarMetrics) waitForClose(run model.Run) {
	<-s.closer
	if s.closed.CompareAndSwap(false, true) {
		m := run.GetModel()
		ssh := fablib.NewSshConfigFactoryImpl(m, s.host.PublicIp)
		if err := fablib.RemoteKill(ssh, "sar"); err != nil {
			logrus.Warnf("did not close sar, it may have already stopped normally (%v)", err)
		}
	}
}

func (s *streamSarMetrics) runSar(ssh fablib.SshConfigFactory) {
	defer func() {
		close(s.joiner)
		logrus.Debugf("joiner closed")
	}()

	for !s.closed.Get() {
		if err := s.reportMetrics(ssh); err != nil {
			return
		}
	}
}

func (s *streamSarMetrics) reportMetrics(ssh fablib.SshConfigFactory) error {
	sar := fmt.Sprintf("sar -u -r -q %d %d", s.intervalSeconds, s.reportIntervalCount)
	output, err := fablib.RemoteExec(ssh, sar)
	if err != nil {
		logrus.Warnf("sar exited (%v)", err)
	}

	summary, err := fablib.SummarizeSar([]byte(output))
	if err != nil {
		logrus.Errorf("sar summary failed (%v) [%s]", err, output)
		return err
	}

	events := summary.ToMetricsEvents()
	m := s.host.GetRegion().GetModel()
	for _, event := range events {
		m.AcceptHostMetrics(s.host, event)
	}

	logrus.Infof("%v sar metrics events reported", len(events))
	return nil
}
