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
	RootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create <model>",
	Short: "create a fablab instance from a model",
	Args:  cobra.ExactArgs(1),
	Run:   create,
}

func create(_ *cobra.Command, args []string) {
	instanceId, err := kernel.NewInstance()
	if err != nil {
		logrus.Fatalf("unable to allocate instance (%w)", err)
	}
	logrus.Infof("allocated new instance [%s]", instanceId)

	modelName := args[0]
	if err := kernel.CreateLabel(instanceId, modelName); err != nil {
		logrus.Fatalf("unable to create instance label [%s] (%w)", instanceId, err)
	}

	_, found := kernel.GetModel(modelName)
	if !found {
		logrus.Fatalf("no model [%s]", modelName)
	}
	logrus.Infof("using model [%s]", modelName)

	if err := kernel.SetActiveInstance(instanceId); err != nil {
		logrus.Fatalf("unable to set active instance (%w)", err)
	}
}
