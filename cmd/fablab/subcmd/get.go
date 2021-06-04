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
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	getCmd.AddCommand(getFilesCmd)
	RootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get entities from remote instances",
}

var getFilesCmd = &cobra.Command{
	Use:   "files <hostSpec> <localPath> <remoteFiles>",
	Short: "copy remote file(s)",
	Args:  cobra.ExactArgs(0),
	Run:   getFiles,
}

func getFiles(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()
	hosts := m.SelectHosts(args[0])
	if len(hosts) != 1 {
		logrus.Fatalf("your hostSpec matched [%d] hosts. must match exactly 1", len(hosts))
	}

	if err := lib.RetrieveRemoteFiles(lib.NewSshConfigFactory(hosts[0]), args[0], args[1:]...); err != nil {
		logrus.Fatalf("error executing remote shell (%v)", err)
	}
}
