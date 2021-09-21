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
	"encoding/json"
	"fmt"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	dumpCmd.AddCommand(dumpHostsCmd)
	RootCmd.AddCommand(dumpCmd)
}

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump the resolved model structure",
	Args:  cobra.ExactArgs(0),
	Run:   dump,
}

func dump(_ *cobra.Command, _ []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()
	if data, err := json.MarshalIndent(m.Dump(), "", "  "); err == nil {
		fmt.Println()
		fmt.Println(string(data))
	} else {
		logrus.Fatalf("error marshaling model dump (%v)", err)
	}
}

var dumpHostsCmd = &cobra.Command{
	Use:   "hosts <host-spec>?",
	Short: "dump the resolved hosts structure",
	Args:  cobra.MaximumNArgs(2),
	Run:   dumpHosts,
}

func dumpHosts(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()

	hostSpec := "*"

	if len(args) > 0 {
		hostSpec = args[0]
	}

	hosts := m.SelectHosts(hostSpec)
	var hostDumps []*model.HostDump
	for _, host := range hosts {
		hostDumps = append(hostDumps, model.DumpHost(host))
	}
	if data, err := json.MarshalIndent(hostDumps, "", "  "); err == nil {
		fmt.Println()
		fmt.Println(string(data))
	} else {
		logrus.Fatalf("error marshaling hosts dump (%v)", err)
	}

}
