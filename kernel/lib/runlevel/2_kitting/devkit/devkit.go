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

package devkit

import (
	"fmt"
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type devKit struct {
	root     func() string
	binaries []string
}

func DevKit(root string, binaries []string) model.Stage {
	return &devKit{root: func() string { return root }, binaries: binaries}
}

func DevKitF(root func() string, binaries []string) model.Stage {
	return &devKit{root: root, binaries: binaries}
}

func (devKit *devKit) Execute(model.Run) error {
	cfgRoot := filepath.Join(model.KitBuild(), "cfg")
	fi, err := os.Stat(model.ConfigBuild())
	if err == nil && fi.IsDir() {
		if err := lib.CopyTree(model.ConfigBuild(), cfgRoot); err != nil {
			return fmt.Errorf("error copying configuration tree (%s)", err)
		}
	} else {
		logrus.Infof("no [cfg] root, not kitting")
	}

	pkiRoot := filepath.Join(model.KitBuild(), "pki")
	fi, err = os.Stat(model.PkiBuild())
	if err == nil && fi.IsDir() {
		if err := lib.CopyTree(model.PkiBuild(), pkiRoot); err != nil {
			return fmt.Errorf("error copying pki tree (%s)", err)
		}
	} else {
		logrus.Infof("no [pki] root, not kitting")
	}

	if err := os.MkdirAll(filepath.Join(model.KitBuild(), "bin"), os.ModePerm); err != nil {
		return fmt.Errorf("error creating kit bin directory (%s)", err)
	}
	for _, binary := range devKit.binaries {
		srcPath := filepath.Join(devKit.root(), binary)
		dstPath := filepath.Join(model.KitBuild(), "bin", binary)
		if _, err := lib.CopyFile(srcPath, dstPath); err == nil {
			logrus.Infof("[%s] => [%s]", srcPath, dstPath)
		} else {
			return fmt.Errorf("error copying binary [%s] => [%s] (%w)", srcPath, dstPath, err)
		}
		if err := os.Chmod(dstPath, 0755); err != nil {
			return fmt.Errorf("error setting binary [%s] permissions (%w)", dstPath, err)
		}
	}
	return nil
}
