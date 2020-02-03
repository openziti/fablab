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

package fablib

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
	"net"
	"os"
	"strconv"
	"strings"
)

func LaunchService(user, host, name, cfg string) error {
	serviceCmd := fmt.Sprintf("nohup /home/%s/fablab/bin/%s --log-formatter pfxlog run /home/%s/fablab/cfg/%s > %s.log 2>&1 &", user, name, user, cfg, name)
	if value, err := RemoteExec(user, host, serviceCmd); err == nil {
		if len(value) > 0 {
			logrus.Infof("output [%s]", strings.Trim(string(value), " \t\r\n"))
		}
	} else {
		return err
	}
	return nil
}

func KillService(user, host, name string) error {
	return RemoteKill(user, host, fmt.Sprintf("/home/%s/fablab/bin/%s", user, name))
}

func RemoteShell(user, host string) error {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			sshAgent(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	logrus.Infof("shell for [%s]", host)

	client, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = session.Close()
		_ = terminal.Restore(fd, oldState)
	}()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		panic(err)
	}

	if err := session.RequestPty("xterm", termHeight, termWidth, ssh.TerminalModes{ssh.ECHO: 1}); err != nil {
		return err
	}

	err = session.Run("/bin/bash")
	if err != nil {
		return err
	}

	return nil
}

func RemoteConsole(user, host, cmd string) error {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			sshAgent(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	logrus.Infof("console for [%s]: '%s'", host, cmd)

	client, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer func() { _ = session.Close() }()

	if err := session.RequestPty("xterm", 40, 80, ssh.TerminalModes{ssh.ECHO: 0}); err != nil {
		return err
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	err = session.Run(cmd)
	if err != nil {
		return err
	}

	return nil
}

func RemoteExec(user, host, cmd string) (string, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			sshAgent(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	logrus.Infof("executing [%s]: '%s'", host, cmd)

	client, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		return "", err
	}

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer func() { _ = session.Close() }()
	var b bytes.Buffer
	session.Stdout = &b

	err = session.Run(cmd)
	if err != nil {
		return "", err
	}

	return b.String(), err
}

func RemoteKill(user, host, match string) error {
	output, err := RemoteExec(user, host, "ps x")
	if err != nil {
		return fmt.Errorf("unable to get remote process listing [%s] (%s)", host, err)
	}

	var pidList []int
	r := strings.NewReader(output)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, match) {
			logrus.Infof("line [%s]", scanner.Text())
			tokens := strings.Split(strings.Trim(line, " \t\n"), " ")
			if pid, err := strconv.Atoi(tokens[0]); err == nil {
				pidList = append(pidList, pid)
			} else {
				return fmt.Errorf("unexpected ps output")
			}
		}
	}

	if len(pidList) > 0 {
		killCmd := "kill"
		for _, pid := range pidList {
			killCmd += fmt.Sprintf(" %d", pid)
		}

		output, err = RemoteExec(user, host, killCmd)
		if err != nil {
			return fmt.Errorf("unable to kill [%s] (%s)", host, err)
		}
	}

	return nil
}

func sshAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}
