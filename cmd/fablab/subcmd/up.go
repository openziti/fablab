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
	"github.com/openziti/fablab/kernel/fablib/figlet"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "progress through lifecycle runlevels (express -> build -> sync -> activate)",
	Args:  cobra.ExactArgs(0),
	Run:   up,
}

func up(_ *cobra.Command, _ []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%v)", err)
	}

	l := model.GetLabel()
	if l == nil {
		logrus.Fatalf("no label for instance [%s]", model.ActiveInstancePath())
	}

	if l != nil {
		m, found := model.GetModel(l.Model)
		if !found {
			logrus.Fatalf("no such model [%s]", l.Model)
		}

		ctx := model.NewRun(l, m)

		figlet.Figlet("infrastructure")

		if err := m.Express(ctx); err != nil {
			logrus.Fatalf("error expressing (%v)", err)
		}

		if err := model.Bootstrap(); err != nil {
			logrus.Fatalf("error re-bootstrapping (%v)", err)
		}

		figlet.Figlet("configuration")

		if err := m.Build(ctx); err != nil {
			logrus.Fatalf("error building (%v)", err)
		}

		figlet.Figlet("distribution")

		if err := m.Sync(ctx); err != nil {
			logrus.Fatalf("error distributing (%v)", err)
		}

		figlet.Figlet("activation")

		if err := m.Activate(ctx); err != nil {
			logrus.Fatalf("error activating (%v)", err)
		}

		figlet.Figlet("FABUL0US!1!")

	} else {
		logrus.Fatalf("no label for run")
	}
}
