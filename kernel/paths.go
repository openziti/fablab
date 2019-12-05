/*
	Copyright 2019 Netfoundry, Inc.

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

package kernel

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func FablabRoot() string {
	return fablabRoot
}

func ConfigSrc() string {
	return filepath.Join(fablabRoot, "lib/templates/cfg")
}

func ConfigBuild() string {
	return filepath.Join(ActiveInstancePath(), "cfg")
}

func KitBuild() string {
	return filepath.Join(ActiveInstancePath(), "kit")
}

func PkiBuild() string {
	return filepath.Join(ActiveInstancePath(), "pki")
}

func configRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("unable to get user home directory (%w)", err)
	}
	return filepath.Join(home, ".fablab")
}

func bootstrapPaths() error {
	fablabRoot = os.Getenv("FABLAB_ROOT")
	if fablabRoot == "" {
		return fmt.Errorf("please set 'FABLAB_ROOT'")
	}
	if fi, err := os.Stat(fablabRoot); err == nil {
		if !fi.IsDir() {
			return fmt.Errorf("invalid 'FABLAB_ROOT' (!directory)")
		}
		logrus.Debugf("FABLAB_ROOT = [%s]", fablabRoot)
	} else {
		return fmt.Errorf("non-existent 'FABLAB_ROOT'")
	}

	return nil
}
