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
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(execCmd)
}

var execCmd = &cobra.Command{
	Use:   "exec <action>",
	Short: "execute an action",
	Args:  cobra.ExactArgs(1),
	Run:   exec,
}

func exec(_ *cobra.Command, args []string) {
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

		action := args[0]
		p, found := m.GetAction(action)
		if !found {
			logrus.Fatalf("no such action [%s]", action)
		}

		if err := p.Execute(m); err != nil {
			logrus.Fatalf("action failed [%s] (%s)", action, err)
		}
	}
}
