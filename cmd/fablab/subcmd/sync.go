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
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(syncCmd)
	syncCmd.AddCommand(syncBinariesCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "synchronize a run kit onto the network",
	Args:  cobra.ExactArgs(0),
	Run:   sync,
}

func sync(_ *cobra.Command, _ []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	ctx, err := model.NewRun()
	if err != nil {
		logrus.WithError(err).Fatal("error initializing run")
	}
	if err := ctx.GetModel().Sync(ctx); err != nil {
		logrus.Fatalf("error synchronizing all hosts (%s)", err)
	}
}

var syncBinariesCmd = &cobra.Command{
	Use:   "binaries",
	Short: "synchronize only the binaries in a run kit onto the network",
	Args:  cobra.ExactArgs(0),
	Run:   syncBinaries,
}

func syncBinaries(_ *cobra.Command, _ []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	ctx, err := model.NewRun()
	if err != nil {
		logrus.WithError(err).Fatal("error initializing run")
	}

	if err := ctx.GetModel().Build(ctx); err != nil {
		logrus.Fatalf("error building configuration (%v)", err)
	}

	ctx.GetModel().Scope.PutVariable("sync.target", "bin")
	if err := ctx.GetModel().Sync(ctx); err != nil {
		logrus.Fatalf("error synchronizing all hosts (%s)", err)
	}
}

var syncConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "synchronize only the config files in a run kit onto the network",
	Args:  cobra.ExactArgs(0),
	Run:   syncBinaries,
}

func syncConfig(_ *cobra.Command, _ []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	ctx, err := model.NewRun()
	if err != nil {
		logrus.WithError(err).Fatal("error initializing run")
	}

	if err := ctx.GetModel().Build(ctx); err != nil {
		logrus.Fatalf("error building configuration (%v)", err)
	}

	ctx.GetModel().Scope.PutVariable("sync.target", "cfg")
	if err := ctx.GetModel().Sync(ctx); err != nil {
		logrus.Fatalf("error synchronizing all hosts (%s)", err)
	}
}
