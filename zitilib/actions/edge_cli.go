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

package zitilib_actions

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/openziti/fablab/kernel/model"
	zitilib_bootstrap "github.com/openziti/fablab/zitilib"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

func Edge(args ...string) model.ActionWithOutput {
	return &edge{
		args: args,
	}
}

func (a *edge) Execute(m *model.Model) error {
	_, err := a.ExecuteWithOutput(m)
	return err
}

func (a *edge) ExecuteWithOutput(m *model.Model) (string, error) {
	if !m.IsBound() {
		return "", errors.New("model not bound")
	}

	allArgs := append(a.args)
	cli := exec.Command(zitilib_bootstrap.ZitiCli(), allArgs...)
	logrus.Infof("executing: %s", zitilib_bootstrap.ZitiCli())
	var cliOut bytes.Buffer
	cli.Stdout = &cliOut
	var cliErr bytes.Buffer
	cli.Stderr = &cliErr
	logrus.Infof("%v", cli.Args)
	err := cli.Run()
	out := fmt.Sprintf("out:[%s], err:[%s]", strings.Trim(cliOut.String(), " \t\r\n"), strings.Trim(cliErr.String(), " \t\r\n"))
	logrus.Info(out)
	if err != nil {
		return "", err
	}
	return cliOut.String(), nil
}

type edge struct {
	args []string
}
