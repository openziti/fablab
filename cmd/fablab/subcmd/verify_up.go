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
	"time"
)

func init() {
	RootCmd.AddCommand(newVerifyUpCmd())
}

func newVerifyUpCmd() *cobra.Command {
	action := &verifyUpAction{}

	var cmd = &cobra.Command{
		Use:   "verify-up <componentSpec> [-c concurrency] [-t timeout]",
		Short: "verifies that the selected components are up and running",
		Args:  cobra.ExactArgs(1),
		Run:   action.run,
	}

	cmd.Flags().IntVarP(&action.concurrency, "concurrency", "c", 10, "Number of components to verify in parallel")
	cmd.Flags().DurationVarP(&action.timeout, "timeout", "t", time.Minute, "Timeout per component")

	return cmd
}

type verifyUpAction struct {
	concurrency int
	timeout     time.Duration
}

func (self *verifyUpAction) run(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	run, err := model.NewRun()
	if err != nil {
		logrus.WithError(err).Fatal("error initializing run")
	}

	if err := component.VerifyUpInParallel(args[0], self.timeout, self.concurrency).Execute(run); err != nil {
		logrus.WithError(err).Fatalf("error verifying components")
	}

	c := run.GetModel().SelectComponents(args[0])
	logrus.Infof("%d components verified running", len(c))
}
