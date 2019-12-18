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
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	createCmd.Flags().StringVarP(&createCmdName, "name", "n", "", "name for the new instance")
	RootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create <model>",
	Short: "create a fablab instance from a model",
	Args:  cobra.ExactArgs(1),
	Run:   create,
}
var createCmdName string

func create(_ *cobra.Command, args []string) {
	var instanceId string
	if createCmdName != "" {
		if err := model.NewNamedInstance(createCmdName); err == nil {
			instanceId = createCmdName
		} else {
			logrus.Fatalf("error creating named instance [%s] (%w)", createCmdName, err)
		}
	} else {
		if id, err := model.NewInstance(); err == nil {
			instanceId = id
		} else {
			logrus.Fatalf("error creating instance (%w)", err)
		}
	}
	logrus.Infof("allocated new instance [%s]", instanceId)

	modelName := args[0]
	if err := model.CreateLabel(instanceId, modelName); err != nil {
		logrus.Fatalf("unable to create instance label [%s] (%w)", instanceId, err)
	}

	_, found := model.GetModel(modelName)
	if !found {
		logrus.Fatalf("no model [%s]", modelName)
	}
	logrus.Infof("using model [%s]", modelName)

	if err := model.SetActiveInstance(instanceId); err != nil {
		logrus.Fatalf("unable to set active instance (%w)", err)
	}
}
