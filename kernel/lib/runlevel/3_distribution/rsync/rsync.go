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

package rsync

import (
	"fmt"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"strings"
)

func Rsync(concurrency int) model.DistributionStage {
	return &rsyncStage{
		concurrency: concurrency,
	}
}

func (rsync *rsyncStage) Distribute(run model.Run) error {
	return run.GetModel().ForEachHost("*", rsync.concurrency, func(host *model.Host) error {
		config := newConfig(run.GetModel(), host.PublicIp)
		if err := synchronizeHost(config); err != nil {
			return fmt.Errorf("error synchronizing host [%s/%s] (%s)", host.GetRegion().GetId(), host.GetId(), err)
		}
		return nil
	})
}

type rsyncStage struct {
	concurrency int
}

func synchronizeHost(config *Config) error {
	if output, err := lib.RemoteExec(config.sshConfigFactory, "mkdir -p /home/ubuntu/fablab"); err == nil {
		if output != "" {
			logrus.Infof("output [%s]", strings.Trim(output, " \t\r\n"))
		}
	} else {
		return err
	}

	if err := rsync(config, model.KitBuild()+"/", fmt.Sprintf("ubuntu@%s:/home/ubuntu/fablab", config.sshConfigFactory.Hostname())); err != nil {
		return fmt.Errorf("rsyncStage failed (%w)", err)
	}

	return nil
}

type Config struct {
	sshBin           string
	sshConfigFactory lib.SshConfigFactory
	rsyncBin         string
}

func newConfig(m *model.Model, publicIp string) *Config {
	config := &Config{
		sshBin:           "ssh",
		sshConfigFactory: lib.NewSshConfigFactoryImpl(m, publicIp),
		rsyncBin:         "rsync",
	}

	if rsyncBin, ok := m.Variables.Must("distribution", "rsync_bin").(string); ok {
		config.rsyncBin = rsyncBin
	}

	if sshBin, ok := m.Variables.Must("distribution", "ssh_bin").(string); ok {
		config.sshBin = sshBin
	}

	return config
}

func (config *Config) sshIdentityFlag() string {
	if config.sshConfigFactory.KeyPath() != "" {
		return "-i " + config.sshConfigFactory.KeyPath()
	}

	return ""
}

func (config *Config) SshCommand() string {
	return config.sshBin + " " + config.sshIdentityFlag()
}
