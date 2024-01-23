/*
	(c) Copyright NetFoundry Inc. Inc.

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

package subcmd

import (
	"fmt"
	"github.com/openziti/fablab/kernel/libssh"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

func init() {
	RootCmd.AddCommand(newSshCmd())
}

type sshCmd struct {
	forceBuiltIn bool
}

func newSshCmd() *cobra.Command {
	cmd := &sshCmd{}

	var cobraCmd = &cobra.Command{
		Use:   "ssh <hostSpec>",
		Short: "establish an ssh connection to a host in the model",
		Args:  cobra.ExactArgs(1),
		Run:   cmd.ssh,
	}

	cobraCmd.Flags().BoolVarP(&cmd.forceBuiltIn, "force-built-in", "f", false,
		"Force use of built-in ssh client, don't try and detect/use an external ssh client")

	return cobraCmd
}

func (self *sshCmd) ssh(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()
	hosts := m.SelectHosts(args[0])
	if len(hosts) != 1 {
		logrus.Fatalf("your regionSpec and hostSpec matched [%d] hosts. must match exactly 1", len(hosts))
	}

	sshCfg := hosts[0].NewSshConfigFactory()

	if !self.forceBuiltIn {
		_, err := exec.LookPath("ssh")
		if err == nil {
			nativeSsh(sshCfg)
			return
		}
	}

	if err := libssh.RemoteShell(sshCfg); err != nil {
		logrus.Fatalf("error executing remote shell (%v)", err)
	}
}

func nativeSsh(sshCfg libssh.SshConfigFactory) {
	cmdArgs := []string{
		"-i", sshCfg.KeyPath(),
		"-o", "StrictHostKeyChecking no",
		sshCfg.User() + "@" + sshCfg.Hostname(),
	}

	if sshCfg.Port() != 22 {
		cmdArgs = append(cmdArgs, "-p", fmt.Sprintf("%v", sshCfg.Port()))
	}

	cmd := exec.Command("ssh", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
