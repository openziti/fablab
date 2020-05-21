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
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "progress through lifecycle runlevels (express -> build -> kit -> sync -> activate)",
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

		fablib.Figlet("infrastructure")

		if err := m.Express(l); err != nil {
			logrus.Fatalf("error expressing (%v)", err)
		}

		if err := model.Bootstrap(); err != nil {
			logrus.Fatalf("error re-bootstrapping (%v)", err)
		}

		fablib.Figlet("configuration")

		if err := m.Build(l); err != nil {
			logrus.Fatalf("error building (%v)", err)
		}

		fablib.Figlet("kitting")

		if err := m.Kit(l); err != nil {
			logrus.Fatalf("error kitting (%v)", err)
		}

		fablib.Figlet("distribution")

		if err := m.Sync(l); err != nil {
			logrus.Fatalf("error distributing (%v)", err)
		}

		fablib.Figlet("activation")

		if err := m.Activate(l); err != nil {
			logrus.Fatalf("error activating (%v)", err)
		}

		fablib.Figlet("FABUL0US!1!")

	} else {
		logrus.Fatalf("no label for run")
	}
}
