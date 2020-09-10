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

package zitilib

import (
	"github.com/openziti/fablab/kernel/fablib/runlevel/2_kitting/devkit"
	"github.com/openziti/fablab/kernel/model"
	"path/filepath"
)

func ZitiRoot() string {
	return zitiRoot
}

func ZitiDistRoot() string {
	if zitiDistRoot == "" {
		return ZitiRoot()
	}
	return zitiDistRoot
}

func zitiBinaries() string {
	return filepath.Join(zitiRoot, "bin")
}

func ZitiDistBinaries() string {
	return filepath.Join(ZitiDistRoot(), "bin")
}

func ZitiCli() string {
	return filepath.Join(zitiBinaries(), "ziti")
}

func ZitiFabricCli() string {
	return filepath.Join(zitiBinaries(), "ziti-fabric")
}

func DefaultZitiBinaries() model.ConfigurationStage {
	zitiBinaries := []string{
		"ziti-controller",
		"ziti-fabric",
		"ziti-fabric-test",
		"ziti-router",
	}

	return devkit.DevKit(ZitiDistBinaries(), zitiBinaries)
}

var zitiRoot string
var zitiDistRoot string
