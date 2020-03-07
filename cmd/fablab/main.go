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

package main

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/netfoundry/fablab/cmd/fablab/subcmd"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/zitilib"
	_ "github.com/netfoundry/fablab/zitilib"
	_ "github.com/netfoundry/fablab/zitilib/models/characterization"
	_ "github.com/netfoundry/fablab/zitilib/models/examples"
	_ "github.com/netfoundry/fablab/zitilib/models/mattermozt"
	"github.com/netfoundry/ziti-foundation/transport"
	"github.com/netfoundry/ziti-foundation/transport/quic"
	"github.com/netfoundry/ziti-foundation/transport/tcp"
	"github.com/netfoundry/ziti-foundation/transport/tls"
	"github.com/sirupsen/logrus"
)

func init() {
	pfxlog.Global(logrus.InfoLevel)
	pfxlog.SetPrefix("github.com/netfoundry/")
	transport.AddAddressParser(quic.AddressParser{})
	transport.AddAddressParser(tls.AddressParser{})
	transport.AddAddressParser(tcp.AddressParser{})
	model.AddBootstrapExtension(&zitilib.Bootstrap{})
}

func main() {
	if err := subcmd.Execute(); err != nil {
		logrus.Fatalf("failure (%s)", err)
	}
}
