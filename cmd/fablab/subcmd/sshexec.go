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
	"bytes"
	"fmt"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"html/template"
	"io"
	"os"
)

func init() {
	cmd := newSshExecCmd()
	RootCmd.AddCommand(cmd.cobraCmd)
}

type sshExecCmd struct {
	cobraCmd    *cobra.Command
	concurrency int
}

func newSshExecCmd() *sshExecCmd {
	cmd := &sshExecCmd{
		cobraCmd: &cobra.Command{
			Use:   "sshexec <hostSpec> <cmd> [<output-file-template>]",
			Short: "establish an ssh connection to the model and runs the given command on the selected hosts",
			Args:  cobra.RangeArgs(2, 3),
		},
	}

	cmd.cobraCmd.Run = cmd.run
	cmd.cobraCmd.Flags().IntVarP(&cmd.concurrency, "concurrency", "c", 1, "Number of hosts to run in parallel")
	return cmd
}

func (cmd *sshExecCmd) run(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()
	logrus.Infof("executing %v with concurrency %v", args[1], cmd.concurrency)
	var tmpl *template.Template
	if len(args) == 3 {
		var err error
		tmpl, err = template.New("output-file-name").Parse(args[2])
		if err != nil {
			logrus.WithError(err).Fatalf("invalid file name template: %v", args[2])
		}
	}

	err := m.ForEachHost(args[0], cmd.concurrency, func(h *model.Host) error {
		var buf *bytes.Buffer
		var out io.Writer
		if tmpl != nil {
			buf := &bytes.Buffer{}
			if err := tmpl.Execute(buf, h); err != nil {
				return err
			}
			fileName := buf.String()
			file, err := os.Create(fileName)
			if err != nil {
				return err
			}
			defer func() { _ = file.Close() }()
			out = file
			logrus.Infof("[%v] output -> %v", h.PublicIp, fileName)
		} else {
			buf = &bytes.Buffer{}
			out = buf
		}
		sshConfigFactory := lib.NewSshConfigFactory(h)
		err := lib.RemoteExecAllTo(sshConfigFactory, out, args[1])
		if err != nil {
			if buf != nil {
				logrus.Errorf("output [%s]", buf.String())
			}
			return fmt.Errorf("error executing process on [%s] (%s)", h.PublicIp, err)
		}
		if buf != nil {
			logrus.Infof("[%v] output:\n%s", h.PublicIp, buf.String())
		}
		return nil
	})

	if err != nil {
		logrus.Fatalf("error executing remote shell (%v)", err)
	}
}
