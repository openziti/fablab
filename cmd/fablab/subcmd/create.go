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

package subcmd

import (
	"fmt"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

func init() {
	RootCmd.AddCommand(NewCreateCommand())
}

func NewCreateCommand() *cobra.Command {
	createCmd := &CreateCommand{}

	cmd := &cobra.Command{
		Use:   "create <model>",
		Short: "create a fablab instance from a model",
		Args:  cobra.MaximumNArgs(1),
		RunE:  createCmd.create,
	}

	cmd.Flags().StringVarP(&createCmd.Name, "name", "n", "", "name for the new instance")
	cmd.Flags().StringVarP(&createCmd.WorkingDir, "directory", "d", "", "working directory for the new instance")
	cmd.Flags().StringVarP(&createCmd.Executable, "executable", "e", "",
		"path to the model specific fablab executable. defaults to the current executable")
	cmd.Flags().StringToStringVarP(&createCmd.Bindings, "label", "l", nil, "label bindings to include in the model")

	return cmd
}

type CreateCommand struct {
	Name       string
	WorkingDir string
	Bindings   map[string]string
	Executable string
}

func (self *CreateCommand) create(*cobra.Command, []string) error {
	if self.Executable == "" {
		executable, err := exec.LookPath(os.Args[0])
		if err != nil {
			return fmt.Errorf("unable to get path (%w)", err)
		}
		self.Executable = executable
	}

	_, err := os.Stat(self.Executable)
	if os.IsNotExist(err) {
		return fmt.Errorf("invalid executable path '%s' (%w)", self.Executable, err)
	}

	if model.GetModel() == nil {
		return errors.New("no model configured, exiting")
	}

	if model.GetModel().GetId() == "" {
		return errors.New("no model id provided, exiting")
	}

	instanceId, err := model.NewInstance(self.Name, self.WorkingDir, self.Executable)
	if err != nil {
		return errors.Wrapf(err, "unable to create instance of model %v, exiting", model.GetModel().Id)
	}

	logrus.Infof("allocated new instance [%v] for model %v", instanceId, model.GetModel().GetId())

	if err := model.CreateLabel(instanceId, self.Bindings); err != nil {
		return errors.Wrapf(err, "unable to create instance label [%s]", instanceId)
	}
	return nil
}
