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

package subcmd

import (
	"bitbucket.org/netfoundry/fablab/kernel"
	"bitbucket.org/netfoundry/fablab/kernel/lib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "execute all lifecycle stages (express -> build -> kit -> sync -> activate)",
	Args:  cobra.ExactArgs(0),
	Run:   up,
}

func up(_ *cobra.Command, _ []string) {
	if err := kernel.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%w)", err)
	}

	l := kernel.GetLabel()
	if l == nil {
		logrus.Fatalf("no label for instance [%s]", kernel.ActiveInstancePath())
	}

	if l != nil {
		m, found := kernel.GetModel(l.Model)
		if !found {
			logrus.Fatalf("no such model [%s]", l.Model)
		}

		lib.Figlet("infrastructure")

		if err := m.Express(l); err != nil {
			logrus.Fatalf("error expressing (%w)", err)
		}

		if err := kernel.Bootstrap(); err != nil {
			logrus.Fatalf("error re-bootstrapping (%w)", err)
		}

		lib.Figlet("configuration")

		if err := m.Build(l); err != nil {
			logrus.Fatalf("error building (%w)", err)
		}

		lib.Figlet("kitting")

		if err := m.Kit(l); err != nil {
			logrus.Fatalf("error kitting (%w)", err)
		}

		lib.Figlet("distribution")

		if err := m.Sync(l); err != nil {
			logrus.Fatalf("error distributing (%w)", err)
		}

		lib.Figlet("activation")

		if err := m.Activate(l); err != nil {
			logrus.Fatalf("error activating (%w)", err)
		}

		lib.Figlet("FABUL0US!1!")

	} else {
		logrus.Fatalf("no label for run")
	}
}
