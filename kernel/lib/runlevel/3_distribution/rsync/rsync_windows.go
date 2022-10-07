/*
	Copyright 2019 NetFoundry Inc.

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
	"github.com/sirupsen/logrus"
	"strings"
)

func rsync(config *Config, sourcePath, targetPath string) error {
	//Only Cygwin's OpenSSH ssh binary and Cygwin's rsync binary used together seem to work.
	//Using cwRsync + Microsoft's OpenSSH ssh port did not work 1st quarter 2020.
	if !strings.Contains(config.rsyncBin, "cygwin") {
		logrus.Warn("on Windows it is highly suggested that Cygwin's rsync binary is used")
	}

	if !strings.Contains(config.sshBin, "cygwin") {
		logrus.Warn("on Windows it is highly suggested that Cygwin's OpenSSH ssh binary is used")
	}

	//rsync at version 3.1.2 on Windows has a 'bug' where if drive letter colons (i.e. the : in C:\) trigger
	//rsync to think that the path is a remote machine. It assumes anything with a colon is a remote machine + path.
	//To work around this, sourcePath should be a directory and we swap into it and use "." or "./" to refer to it
	rsync := lib.NewProcess(config.rsyncBin, "-avz", "-e", config.SshCommand()+` -o StrictHostKeyChecking=no`, "--delete", ".", targetPath)
	rsync.Cmd.Dir = sourcePath

	rsync.WithTail(lib.StdoutTail)
	if err := rsync.Run(); err != nil {
		return fmt.Errorf("rsync failed (%w)", err)
	}
	return nil
}
