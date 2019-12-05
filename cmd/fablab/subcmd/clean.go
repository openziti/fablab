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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(cleanCmd)
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "remove instance data from empty or disposed models",
	Args:  cobra.ExactArgs(0),
	Run:   clean,
}

func clean(_ *cobra.Command, _ []string) {
	if err := kernel.BootstrapInstance(); err != nil {
		logrus.Fatalf("error bootstrapping instance (%w)", err)
	}

	instanceIds, err := kernel.ListInstances()
	if err != nil {
		logrus.Fatalf("error listing instances (%w)", err)
	}

	activeInstanceId := kernel.ActiveInstanceId()
	for _, instanceId := range instanceIds {
		if l, err := kernel.LoadLabelForInstance(instanceId); err == nil {
			if l.State == kernel.Created || l.State == kernel.Disposed {
				if err := kernel.RemoveInstance(instanceId); err != nil {
					logrus.Fatalf("error removing instance [%s] (%w)", instanceId, err)
				}
				if instanceId == activeInstanceId {
					if err := kernel.ClearActiveInstance(); err != nil {
						logrus.Errorf("error clearing active instance (%w)", err)
					}
				}
				logrus.Infof("removed instance [%s]", instanceId)
			}
		} else {
			logrus.Warnf("error loading label for instance [%s] (%w)", instanceId, err)
		}
	}
}
