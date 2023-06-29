/*
	Copyright 2019 NetFoundry Inc.

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
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	RootCmd.AddCommand(verifyUpCmd)
}

var verifyUpCmd = &cobra.Command{
	Use:   "verify-up <componentSpec>",
	Short: "verifies that the selected components are up and running",
	Args:  cobra.ExactArgs(1),
	Run:   verifyUp,
}

func verifyUp(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()
	components := m.SelectComponents(args[0])
	if len(components) == 0 {
		logrus.Fatal("your component spec matched 0 components, it should match at least 1")
	}

	run, err := model.NewRun()
	if err != nil {
		logrus.WithError(err).Fatal("error initializing run")
	}

	deadline := time.Now().Add(time.Minute)

	for _, c := range components {
		log := pfxlog.Logger().WithField("componentId", c.Id)

		running := false
		for !running {
			if time.Now().After(deadline) {
				log.Fatal("timed out waiting for component to be running")
			}
			running, err = c.IsRunning(run)
			if err != nil {
				log.WithError(err).Fatal("unable to check component status")
			}
			if !running {
				log.Info("component not running yet, waiting 1 second")
			}
			time.Sleep(time.Second)
		}
		log.Info("component is running")
	}
}
