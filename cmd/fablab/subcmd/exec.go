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
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	execCmd.Flags().StringArrayVarP(&execCmdBindings, "variable", "b", []string{}, "specify variable binding ('<hostSpec>.a.b.c=value')")
	RootCmd.AddCommand(execCmd)
}

var execCmd = &cobra.Command{
	Use:   "exec <action>",
	Short: "execute an action",
	Args:  cobra.ExactArgs(1),
	Run:   exec,
}
var execCmdBindings []string

func exec(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()

	if !m.IsBound() {
		logrus.Fatalf("model not bound")
	}

	for _, binding := range execCmdBindings {
		if err := execCmdBind(m, binding); err != nil {
			logrus.Fatalf("error binding [%s] (%v)", binding, err)
		}
	}

	action := args[0]
	p, found := m.GetAction(action)
	if !found {
		logrus.Fatalf("no such action [%s]", action)
	}

	if err := p.Execute(m); err != nil {
		logrus.Fatalf("action failed [%s] (%s)", action, err)
	}
}

func execCmdBind(m *model.Model, binding string) error {
	halves := strings.Split(binding, "=")
	if len(halves) != 2 {
		return errors.New("variable path and value must be separated by '='")
	}
	path := strings.Split(halves[0], ".")
	if len(path) < 2 {
		return errors.New("path must be of form <hostSpec>.v1...=")
	}
	host, err := m.SelectHost(path[0])
	if err != nil {
		return errors.Wrap(err, "missing host")
	}

	key := strings.Join(path[1:], ".")
	host.PutVariable(key, halves[1])
	return nil
}
