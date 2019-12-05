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
	"github.com/netfoundry/fablab/kernel"
	"github.com/netfoundry/fablab/kernel/lib"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

func Rsync() kernel.DistributionStage {
	return &rsyncStage{}
}

func (rsync *rsyncStage) Distribute(m *kernel.Model) error {
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

func synchronizeHost(h *kernel.Host, sshUsername string) error {
	if output, err := lib.RemoteExec(sshUsername, h.PublicIp, "mkdir -p /home/fedora/fablab"); err == nil {
		if output != "" {
			logrus.Infof("output [%s]", strings.Trim(output, " \t\r\n"))
		}
	} else {
		return err
	}

	if err := rsync(kernel.KitBuild()+"/", fmt.Sprintf("fedora@%s:/home/fedora/fablab", h.PublicIp)); err != nil {
		return fmt.Errorf("rsyncStage failed (%w)", err)
	}

	return nil
}

func rsync(sourcePath, targetPath string) error {
	rsync := lib.NewProcess("rsync", "-avz", "-e", "ssh -o \"StrictHostKeyChecking no\"", "--delete", sourcePath, targetPath)
	rsync.WithTail(lib.StdoutTail)
	if err := rsync.Run(); err != nil {
		return fmt.Errorf("rsync failed (%w)", err)
	}
	return nil
}