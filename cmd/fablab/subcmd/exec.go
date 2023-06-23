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
	Use:   "exec <action> [<actions>...]",
	Short: "execute one or more actions",
	Args:  cobra.MinimumNArgs(1),
	Run:   exec,
}
var execCmdBindings []string

func exec(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	ctx, err := model.NewRun()
	if err != nil {
		logrus.WithError(err).Fatal("error initializing run")
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

	var actions []model.Action

	for _, name := range args {
		action, found := m.GetAction(name)
		if !found {
			logrus.Fatalf("no such action [%s]", name)
		}
		actions = append(actions, action)
	}

	for _, action := range actions {
		if err := action.Execute(ctx); err != nil {
			logrus.WithError(err).Fatalf("action failed [%s]", action)
		}
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
