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
	"fmt"
	"github.com/michaelquigley/pfxlog"
	terraform0 "github.com/openziti/fablab/kernel/lib/runlevel/0_infrastructure/terraform"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	RootCmd.AddCommand(NewImportCommand().Command)
}

func NewImportCommand() *ImportCommand {
	result := &ImportCommand{
		Command: &cobra.Command{
			Use:   "import instance",
			Short: "adds an instance to the config and sets it as default",
			Args:  cobra.ExactArgs(1),
		},
	}

	result.Command.RunE = result.runImport

	return result
}

type ImportCommand struct {
	*cobra.Command
}

func (self *ImportCommand) runImport(_ *cobra.Command, args []string) error {
	p := args[0]
	fi, err := os.Stat(p)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return self.importDir(p)
	}

	isTgzTmp := false
	tgzFile := p
	if strings.HasSuffix(p, "gpg") {
		tgzFile = strings.TrimSuffix(p, ".gpg")
		isTgzTmp = true
		gpgArgs := []string{"--quiet", "--batch", "--yes", "--decrypt",
			fmt.Sprintf("--passphrase=%s", os.Getenv("FABLAB_PASSPHRASE")),
			"--output", tgzFile, p}
		fmt.Printf("running: gpg %+v\n", gpgArgs)
		command := exec.Command("gpg", gpgArgs...)
		command.Stderr = os.Stderr
		command.Stdout = os.Stdout
		if err = command.Run(); err != nil {
			return err
		}
	}

	if !strings.HasSuffix(tgzFile, ".tar.gz") {
		return errors.New("expecting .tar.gz archive")
	}

	command := exec.Command("tar", "-xzf", tgzFile, "-C", model.UserInstanceRoot())
	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	if err = command.Run(); err != nil {
		return err
	}

	if isTgzTmp {
		if err = os.Remove(tgzFile); err != nil {
			fmt.Printf("failed to delete tmp file [%s]\n", tgzFile)
		}
	}

	instancePath := strings.TrimSuffix(filepath.Base(tgzFile), ".tar.gz")
	instancePath = filepath.Join(model.UserInstanceRoot(), instancePath)
	return self.importDir(instancePath)
}

func (self *ImportCommand) importDir(path string) error {
	instanceDir, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	cfg, err := model.LoadConfig(filepath.Join(instanceDir, "config.yml"))
	if err != nil {
		return errors.Wrap(err, "unable to load config.yml from instance being imported")
	}

	logrus.Infof("attempting to import instance: %s", cfg.Default)
	instance, ok := cfg.Instances[cfg.Default]
	if !ok {
		return errors.Errorf("instance %s not found in config", cfg.Default)
	}

	localConfig := model.GetConfig()

	localConfig.Instances[instance.Id] = &model.InstanceConfig{
		Id:               instance.Id,
		Model:            instance.Model,
		WorkingDirectory: instanceDir,
	}
	localConfig.Default = instance.Id
	if err = model.PersistConfig(localConfig); err != nil {
		return errors.Wrap(err, "unable to persist changes to local config")
	}

	if err = updateWorkingPath(instance.WorkingDirectory, instanceDir); err != nil {
		return err
	}

	logrus.Infof("imported instance at %s, id=[%s], model=[%s]", instanceDir, instance.Id, instance.Model)

	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	_, err = model.NewRun()
	if err != nil {
		logrus.WithError(err).Fatal("error initializing run")
	}

	tf := &terraform0.Terraform{}
	return tf.Init()
}

func updateWorkingPath(oldPath, newPath string) error {
	return fs.WalkDir(os.DirFS("/"), strings.TrimPrefix(newPath, "/"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tf") {
			fullPath := filepath.Join("/", path)
			contents, err := os.ReadFile(fullPath)
			if err != nil {
				return err
			}
			oldContents := string(contents)
			newContents := strings.ReplaceAll(oldContents, oldPath, newPath)
			if oldContents != newContents {
				if err = os.WriteFile(fullPath, []byte(newContents), d.Type()); err != nil {
					return err
				}
				pfxlog.Logger().Infof("rewrote %s", path)
			}
		}
		return nil
	})
}
