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
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"os"
	"strings"
)

func Retrieve(region, host, path, extension string) model.OperatingStage {
	return &retrieve{
		region:    region,
		host:      host,
		path:      path,
		extension: extension,
	}
}

func (self *retrieve) Operate(m *model.Model, run string) error {
	hosts := m.GetHosts(self.region, self.host)
	if len(hosts) == 1 {
		ssh := fablib.NewSshConfigFactoryImpl(m, hosts[0].PublicIp)

		if files, err := fablib.RemoteFileList(ssh, self.path); err == nil {
			paths := make([]string, 0)
			for _, file := range files {
				if strings.HasSuffix(file.Name(), self.extension) {
					paths = append(paths, file.Name())
				}
			}
			forensicsPath := model.AllocateForensicScenario(run, self.region)
			if err := os.MkdirAll(forensicsPath, os.ModePerm); err != nil {
				return fmt.Errorf("error creating forensics root [%s] (%w)", forensicsPath, err)
			}
			if err := fablib.RetrieveRemoteFiles(ssh, forensicsPath, paths...); err != nil {
				return fmt.Errorf("error retrieving remote files (%w)", err)
			}
			if err := fablib.DeleteRemoteFiles(ssh, paths...); err != nil {
				return fmt.Errorf("error deleting remote files (%w)", err)
			}

		} else {
			return fmt.Errorf("error listing remote directory (%w)", err)
		}

	} else {
		return fmt.Errorf("found [%d] hosts", len(hosts))
	}

	return nil
}

type retrieve struct {
	region    string
	host      string
	path      string
	extension string
}
