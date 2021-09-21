/*
	Copyright NetFoundry, Inc.

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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	ipCmd.Flags().BoolVarP(&privateIp, "private", "p", false, "retrieve private ip (not public)")
	RootCmd.AddCommand(ipCmd)
}

var ipCmd = &cobra.Command{
	Use:   "ip <hostSpec>",
	Short: "retrieve an ip address from the model",
	Args:  cobra.ExactArgs(1),
	Run:   ip,
}
var privateIp bool

func ip(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%v)", err)
	}

	m := model.GetModel()

	hosts := m.SelectHosts(args[0])
	for _, host := range hosts {
		if !privateIp {
			fmt.Println(host.PublicIp)
		} else {
			fmt.Println(host.PrivateIp)
		}
	}
}
