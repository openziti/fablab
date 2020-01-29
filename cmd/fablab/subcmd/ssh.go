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

package subcmd

import (
	"github.com/netfoundry/fablab/kernel/actions/host"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(sshCmd)
}

var sshCmd = &cobra.Command{
	Use:   "ssh <regionSpec> <hostSpec>",
	Short: "establish an ssh connection to the model",
	Args:  cobra.ExactArgs(2),
	Run:   ssh,
}

func ssh(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	label := model.GetLabel()
	if label == nil {
		logrus.Fatalf("no label for instance [%s]", model.ActiveInstancePath())
	}

	if label != nil {
		m, found := model.GetModel(label.Model)
		if !found {
			logrus.Fatalf("no such model [%s]", label.Model)
		}

		if !m.IsBound() {
			logrus.Fatalf("model not bound")
		}

		hosts := m.GetHosts(args[0], args[1])
		if len(hosts) != 1 {
			logrus.Fatalf("your regionSpec and hostSpec matched [%d] hosts. must match exactly 1", len(hosts))
		}

		sshUsername := m.MustVariable("credentials", "ssh", "username").(string)
		sshKeyPath := m.Variable("credentials", "ssh", "key_path").(string)
		if err := host.RemoteShell(sshUsername, hosts[0].PublicIp, sshKeyPath); err != nil {
			logrus.Fatalf("error executing remote shell (%w)", err)
		}
	}
}
