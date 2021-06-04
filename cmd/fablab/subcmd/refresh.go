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
	"github.com/openziti/fablab/kernel/lib/figlet"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(refreshCmd)
}

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "progress through lifecycle runlevels (build -> sync -> activate)",
	Args:  cobra.ExactArgs(0),
	Run:   refresh,
}

func refresh(_ *cobra.Command, _ []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%v)", err)
	}

	ctx := model.NewRun()

	figlet.Figlet("configuration")

	if err := ctx.GetModel().Build(ctx); err != nil {
		logrus.Fatalf("error building (%v)", err)
	}

	figlet.Figlet("distribution")

	if err := ctx.GetModel().Sync(ctx); err != nil {
		logrus.Fatalf("error distributing (%v)", err)
	}

	figlet.Figlet("activation")

	if err := ctx.GetModel().Activate(ctx); err != nil {
		logrus.Fatalf("error activating (%v)", err)
	}

	figlet.Figlet("FABUL0US!1!")
}
