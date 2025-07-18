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
	"github.com/openziti/fablab/kernel/lib/actions/component"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(newRestartCmd())
}

func newRestartCmd() *cobra.Command {
	action := &restartAction{}

	var cmd = &cobra.Command{
		Use:   "restart <component-spec> [-c concurrency]",
		Short: "restart components",
		Args:  cobra.ExactArgs(1),
		Run:   action.run,
	}

	cmd.Flags().IntVarP(&action.concurrency, "concurrency", "c", 10, "Number of components to restart in parallel")

	return cmd
}

type restartAction struct {
	concurrency int
}

func (self *restartAction) run(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	ctx, err := model.NewRun()
	if err != nil {
		logrus.WithError(err).Fatal("error initializing run")
	}

	if err = component.StopInParallel(args[0], self.concurrency).Execute(ctx); err != nil {
		logrus.WithError(err).Fatalf("error stopping components")
	}

	if err = component.StartInParallel(args[0], self.concurrency).Execute(ctx); err != nil {
		logrus.WithError(err).Fatalf("error starting components")
	}

	c := ctx.GetModel().SelectComponents(args[0])
	logrus.Infof("%d components restarted", len(c))
}
