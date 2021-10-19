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
	RootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "show the environment and active instance status",
	Args:  cobra.ExactArgs(0),
	Run:   status,
}

func status(_ *cobra.Command, _ []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatal("unable to bootstrap (%w)", err)
	}

	l := model.GetLabel()
	if l == nil {
		fmt.Printf("%-20s no label\n", "Label")
	} else {
		fmt.Printf("%-20s\n", "Label")
		fmt.Printf("%-20s %s\n", "  Model", l.Model)
		fmt.Printf("%-20s %s\n", "  State", l.State)
	}
	fmt.Println()
}
