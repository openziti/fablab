/*
	(c) Copyright NetFoundry Inc. Inc.

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

package main

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/cmd/fablab/subcmd"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

func init() {
	pfxlog.GlobalInit(logrus.InfoLevel, pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti/"))
}

func main() {
	if len(os.Args) > 1 {
		runLocalBinary := false
		switch os.Args[1] {
		case "completion", "clean", "pin", "unpin":
			runLocalBinary = true
		}
		if !runLocalBinary && len(os.Args) > 2 {
			if os.Args[1] == "list" && os.Args[2] == "instances" {
				runLocalBinary = true
			}
		}

		if runLocalBinary {
			if err := subcmd.Execute(); err != nil {
				logrus.Fatalf("failure (%v)", err)
			}
			return
		}
	}

	cfg := model.GetConfig()
	selectedId := cfg.GetSelectedInstanceId()
	instance, ok := cfg.Instances[selectedId]
	if !ok {
		logrus.Fatalf("invalid selected instance '%s'", selectedId)
		return
	}

	if instance.Executable == "" {
		logrus.Fatalf("selected instance '%s' has no executable configured to delegate to", selectedId)
		return
	}

	cmd := exec.Command(instance.Executable, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stderr
	_ = cmd.Run()
}
