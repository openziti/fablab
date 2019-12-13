/*
	Copyright 2019 Netfoundry, Inc.

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
	"github.com/netfoundry/fablab/kernel/internal"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"strings"
)

func Rsync() model.DistributionStage {
	return &rsyncStage{}
}

func (rsync *rsyncStage) Distribute(m *model.Model) error {
	sshUsername := m.MustVariable("credentials", "ssh", "username").(string)
	for regionId, r := range m.Regions {
		for hostId, host := range r.Hosts {
			if err := synchronizeHost(host, sshUsername); err != nil {
				return fmt.Errorf("error synchronizing host [%s/%s] (%s)", regionId, hostId, err)
			}
		}
	}
	return nil
}

type rsyncStage struct {
}

func synchronizeHost(h *model.Host, sshUsername string) error {
	if output, err := internal.RemoteExec(sshUsername, h.PublicIp, "mkdir -p /home/fedora/fablab"); err == nil {
		if output != "" {
			logrus.Infof("output [%s]", strings.Trim(output, " \t\r\n"))
		}
	} else {
		return err
	}

	if err := rsync(model.KitBuild()+"/", fmt.Sprintf("fedora@%s:/home/fedora/fablab", h.PublicIp)); err != nil {
		return fmt.Errorf("rsyncStage failed (%w)", err)
	}

	return nil
}

func rsync(sourcePath, targetPath string) error {
	rsync := internal.NewProcess("rsync", "-avz", "-e", "ssh -o \"StrictHostKeyChecking no\"", "--delete", sourcePath, targetPath)
	rsync.WithTail(internal.StdoutTail)
	if err := rsync.Run(); err != nil {
		return fmt.Errorf("rsync failed (%w)", err)
	}
	return nil
}
