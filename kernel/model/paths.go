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

package model

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

const (
	BuildConfigDir = "cfg"
	BuildKitDir    = "kit"
	BuildPkiDir    = "pki"
	BuildBinDir    = "bin"
	BuildTmpDir    = "tmp"
)

func ScriptBuild() string {
	return MakeBuildPath(BuildBinDir)
}

func ConfigBuild() string {
	return MakeBuildPath(BuildConfigDir)
}

func KitBuild() string {
	return MakeBuildPath("kit")
}

func PkiBuild() string {
	return MakeBuildPath(BuildPkiDir)
}

func MakeBuildPath(path string) string {
	return filepath.Join(BuildPath(), path)
}

func BuildPath() string {
	return instanceConfig.WorkingDirectory
}

func configRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("unable to get user home directory (%v)", err)
	}
	return filepath.Join(home, ".fablab")
}
