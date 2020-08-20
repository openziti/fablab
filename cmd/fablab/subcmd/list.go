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
	"fmt"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(listInstancesCmd)
	listCmd.AddCommand(listModelsCmd)
	listCmd.AddCommand(listHostsCmd)
	RootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list objects",
}

var listInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "list instances",
	Args:  cobra.ExactArgs(0),
	Run:   listInstances,
}

func listInstances(_ *cobra.Command, _ []string) {
	if err := model.BootstrapInstance(); err != nil {
		logrus.Fatalf("unable to bootstrap config (%v)", err)
	}

	activeInstanceId := model.ActiveInstanceId()
	instanceIds, err := model.ListInstances()
	if err != nil {
		logrus.Fatalf("unable to list instances (%v)", err)
	}

	fmt.Println()
	fmt.Printf("[%d] instances:\n\n", len(instanceIds))
	for _, instanceId := range instanceIds {
		idLabel := instanceId
		if instanceId == activeInstanceId {
			idLabel += "*"
		}
		if l, err := model.LoadLabelForInstance(instanceId); err == nil {
			fmt.Printf("%-12s %-24s [%s]\n", idLabel, l.Model, l.State)
		} else {
			fmt.Printf("%-12s %s\n", idLabel, err)
		}
	}
	if len(instanceIds) > 0 {
		fmt.Println()
	}
}

var listModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "list available models",
	Args:  cobra.ExactArgs(0),
	Run:   listModels,
}

func listModels(_ *cobra.Command, _ []string) {
	models := model.ListModels()
	fmt.Printf("\nfound [%d] models:\n\n", len(models))
	for _, modelName := range models {
		fmt.Printf("\t" + modelName + "\n")
	}
	fmt.Println()
}
