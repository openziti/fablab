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

package pki

import (
	"bytes"
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/zitilib/development/bootstrap"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

func generateCa() error {
	pki := exec.Command(zitilib_bootstrap.ZitiCli(), "pki", "create", "ca", "--pki-root", model.PkiBuild(), "--ca-name", "root", "--ca-file", "root")
	var pkiOut bytes.Buffer
	pki.Stdout = &pkiOut
	var pkiErr bytes.Buffer
	pki.Stderr = &pkiErr
	logrus.Infof("%v", pki.Args)
	if err := pki.Run(); err != nil {
		logrus.Errorf("stdOut [%s], stdErr [%s]", strings.Trim(pkiOut.String(), " \t\r\n"), strings.Trim(pkiErr.String(), " \t\r\n"))
		return fmt.Errorf("error generating key (%s)", err)
	}

	pki = exec.Command(zitilib_bootstrap.ZitiCli(), "pki", "create", "intermediate", "--pki-root", model.PkiBuild(), "--ca-name", "root")
	pkiOut.Reset()
	pki.Stdout = &pkiOut
	pkiErr.Reset()
	pki.Stderr = &pkiErr
	logrus.Infof("%v", pki.Args)
	if err := pki.Run(); err != nil {
		logrus.Errorf("stdOut [%s], stdErr [%s]", strings.Trim(pkiOut.String(), " \t\r\n"), strings.Trim(pkiErr.String(), " \t\r\n"))
		return fmt.Errorf("error generating key (%s)", err)
	}

	return nil
}

func generateCert(name, ip string) error {
	logrus.Infof("generating certificate [%s:%s]", name, ip)
	pki := exec.Command(zitilib_bootstrap.ZitiCli(), "pki", "create", "key", "--pki-root", model.PkiBuild(), "--ca-name", "intermediate", "--key-file", name)
	var pkiOut bytes.Buffer
	pki.Stdout = &pkiOut
	var pkiErr bytes.Buffer
	pki.Stderr = &pkiErr
	logrus.Infof("%v", pki.Args)
	if err := pki.Run(); err != nil {
		logrus.Errorf("stdOut [%s], stdErr [%s]", strings.Trim(pkiOut.String(), " \t\r\n"), strings.Trim(pkiErr.String(), " \t\r\n"))
		return fmt.Errorf("error generating key (%s)", err)
	}

	pki = exec.Command(zitilib_bootstrap.ZitiCli(), "pki", "create", "server", "--pki-root", model.PkiBuild(), "--ca-name", "intermediate", "--server-file", fmt.Sprintf("%s-server", name), "--ip", ip, "--key-file", name)
	pkiOut.Reset()
	pki.Stdout = &pkiOut
	pkiErr.Reset()
	pki.Stderr = &pkiErr
	logrus.Infof("%v", pki.Args)
	if err := pki.Run(); err != nil {
		logrus.Errorf("stdOut [%s], stdErr [%s]", strings.Trim(pkiOut.String(), " \t\r\n"), strings.Trim(pkiErr.String(), " \t\r\n"))
		return fmt.Errorf("error generating server certificate (%s)", err)
	}

	pki = exec.Command(zitilib_bootstrap.ZitiCli(), "pki", "create", "client", "--pki-root", model.PkiBuild(), "--ca-name", "intermediate", "--client-file", fmt.Sprintf("%s-client", name), "--key-file", name, "--client-name", name)
	pkiOut.Reset()
	pki.Stdout = &pkiOut
	pkiErr.Reset()
	pki.Stderr = &pkiErr
	logrus.Infof("%v", pki.Args)
	if err := pki.Run(); err != nil {
		logrus.Errorf("stdOut [%s], stdErr [%s]", strings.Trim(pkiOut.String(), " \t\r\n"), strings.Trim(pkiErr.String(), " \t\r\n"))
		return fmt.Errorf("error generating client certificate (%s)", err)
	}

	return nil
}
