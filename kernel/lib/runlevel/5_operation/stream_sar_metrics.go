/*
	Copyright 2020 NetFoundry Inc.

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
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"sync/atomic"
)

func StreamSarMetrics(host *model.Host, intervalSeconds, reportIntervalCount int, runPhase Phase, cleanupPhase Phase) model.Stage {
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
	closed              atomic.Bool
}

func (s *streamSarMetrics) Execute(model.Run) error {
	go s.waitForClose()
	ssh := lib.NewSshConfigFactory(s.host)
	go s.runSar(ssh)
	return nil
}

func (s *streamSarMetrics) waitForClose() {
	<-s.closer
	if s.closed.CompareAndSwap(false, true) {
		ssh := lib.NewSshConfigFactory(s.host)
		if err := lib.RemoteKill(ssh, "sar"); err != nil {
			logrus.Warnf("did not close sar, it may have already stopped normally (%v)", err)
		}
	}
}

func (s *streamSarMetrics) runSar(ssh lib.SshConfigFactory) {
	defer func() {
		close(s.joiner)
		logrus.Debugf("joiner closed")
	}()

	for !s.closed.Load() {
		if err := s.reportMetrics(ssh); err != nil {
			return
		}
	}
}

func (s *streamSarMetrics) reportMetrics(ssh lib.SshConfigFactory) error {
	log := pfxlog.Logger().WithField("addr", ssh.Address())
	sar := fmt.Sprintf("sar -u -r -q %d %d", s.intervalSeconds, s.reportIntervalCount)
	output, err := lib.RemoteExec(ssh, sar)
	if err != nil {
		log.WithError(err).Warn("sar exited")
	}

	summary, err := lib.SummarizeSar([]byte(output))
	if err != nil {
		log.WithError(err).Errorf("sar summary failed [%s]", output)
		return err
	}

	events := summary.ToMetricsEvents()
	m := s.host.GetRegion().GetModel()
	for _, event := range events {
		m.AcceptHostMetrics(s.host, event)
	}

	log.Infof("%v sar metrics events reported", len(events))
	return nil
}
