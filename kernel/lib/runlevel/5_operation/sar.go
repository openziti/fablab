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
	"encoding/json"
	"fmt"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Sar(scenario string, host *model.Host, intervalSeconds int, joiner chan struct{}) model.OperatingStage {
	return &sar{
		scenario:        scenario,
		host:            host,
		intervalSeconds: intervalSeconds,
		joiner:          joiner,
	}
}

func SarCloser(host *model.Host) model.OperatingStage {
	return &sarCloser{
		host: host,
	}
}

func (s *sar) Operate(run model.Run) error {
	m := run.GetModel()
	ssh := lib.NewSshConfigFactoryImpl(m, s.host.PublicIp)
	go s.runSar(ssh)
	return nil
}

func (s *sarCloser) Operate(run model.Run) error {
	m := run.GetModel()
	ssh := lib.NewSshConfigFactoryImpl(m, s.host.PublicIp)
	if err := lib.RemoteKill(ssh, "sar"); err != nil {
		return fmt.Errorf("error closing sar (%w)", err)
	}
	return nil
}

func (s *sar) runSar(ssh lib.SshConfigFactory) {
	defer func() {
		close(s.joiner)
		logrus.Debugf("joiner closed")
	}()

	sar := fmt.Sprintf("sar -u -r -q %d", s.intervalSeconds)
	output, err := lib.RemoteExec(ssh, sar)
	if err != nil {
		logrus.Warnf("sar exited (%v)", err)
	}

	summary, err := lib.SummarizeSar([]byte(output))
	if err != nil {
		logrus.Errorf("sar summary failed (%v) [%s]", err, output)
		return
	}
	j, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		logrus.Errorf("error marshaling summary (%v)", err)
		return
	}
	logrus.Debugf("summary = %s", j)

	if s.host.Data == nil {
		s.host.Data = make(model.Data)
	}
	v, found := s.host.Data["host"]
	if !found {
		v = make(model.Data)
		s.host.Data["host"] = v
	}
	v.(model.Data)[s.scenario] = summary

	logrus.Infof("sar data added to host")
}

type sar struct {
	scenario        string
	host            *model.Host
	intervalSeconds int
	joiner          chan struct{}
}

type sarCloser struct {
	host *model.Host
}
