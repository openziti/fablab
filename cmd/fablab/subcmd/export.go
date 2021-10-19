/*
	Copyright 2020 NetFoundry, Inc.

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
	"github.com/openziti/fablab/kernel/lib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/foundation/util/info"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path/filepath"
)

func init() {
	RootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export the instance data to a zip archive",
	Run:   export,
}

func export(_ *cobra.Command, _ []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()

	zipName := fmt.Sprintf("%s-%d.zip", filepath.Base(model.BuildPath()), info.NowInMilliseconds())

	if err := lib.Export(zipName, m); err != nil {
		logrus.Fatalf("error exporting (%v)", err)
	}
}
