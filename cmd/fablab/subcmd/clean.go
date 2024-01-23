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
	RootCmd.AddCommand(cleanCmd)
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "remove instance data from empty or disposed models",
	Args:  cobra.ExactArgs(0),
	Run:   clean,
}

func clean(_ *cobra.Command, _ []string) {
	if err := model.BootstrapInstance(); err != nil {
		logrus.Fatalf("error bootstrapping instance (%v)", err)
	}

	cfg := model.GetConfig()

	activeInstanceId := model.ActiveInstanceId()
	for instanceId, instanceConfig := range cfg.Instances {
		if l, err := instanceConfig.LoadLabel(); err == nil {
			if l.State == model.Created || l.State == model.Disposed {
				if err := instanceConfig.CleanupWorkingDir(); err != nil {
					logrus.WithError(err).Fatalf("error removing instance [%s]", instanceId)
				}
				if instanceId == activeInstanceId {
					cfg.Default = ""
				}
				delete(cfg.Instances, instanceId)
				if err := model.PersistConfig(cfg); err != nil {
					logrus.WithError(err).Fatalf("error removing instance (%v)", err)
				} else {
					logrus.Infof("removed instance [%s]", instanceId)
				}
			}
		} else {
			logrus.Warnf("error loading label for instance [%s] (%v)", instanceId, err)
		}
	}
}
