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
	"bytes"
	"github.com/openziti/fablab/kernel/libssh"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"text/template"
)

func init() {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "get entities from remote instances",
	}

	getCmd.AddCommand(newGetFilesCmd())
	RootCmd.AddCommand(getCmd)
}

func newGetFilesCmd() *cobra.Command {
	action := &getFilesAction{}
	cmd := &cobra.Command{
		Use:   "files <hostSpec> <localPath> <remoteFiles>",
		Short: "copy remote file(s)",
		Args:  cobra.MinimumNArgs(3),
		RunE:  action.run,
	}
	cmd.Flags().Uint8VarP(&action.concurrency, "concurrency", "c", 1, "How many files to retrieve concurrently")
	return cmd
}

type getFilesAction struct {
	concurrency uint8
}

func (self *getFilesAction) run(_ *cobra.Command, args []string) error {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()
	hosts := m.SelectHosts(args[0])
	if len(hosts) == 0 {
		logrus.Fatalf("your hostSpec matched [%d] hosts. must match at least 1", len(hosts))
	}

	if self.concurrency < 1 {
		logrus.Fatalf("concurrency must be at least 1")
	}

	return m.ForEachHost(args[0], int(self.concurrency), func(host *model.Host) error {
		localPath := args[1]
		tmpl := template.New("localPath")
		tmpl, err := tmpl.Parse(localPath)
		if err != nil {
			return errors.Wrapf(err, "unable to parse template for destination directory '%s'", localPath)
		}
		buf := bytes.NewBuffer(nil)
		if err = tmpl.Execute(buf, host); err != nil {
			return errors.Wrapf(err, "unable to execute template for destination directory '%s'", localPath)
		}
		localPath = buf.String()
		return libssh.RetrieveRemoteFiles(host.NewSshConfigFactory(), localPath, args[2:]...)
	})
}
