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

package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/foundation/v2/info"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func LaunchService(factory SshConfigFactory, name, cfg string, sudo bool) error {
	sudoCmd := ""
	if sudo {
		sudoCmd = " sudo "
	}
	logName := strings.ReplaceAll(name, " ", "-")
	serviceCmd := fmt.Sprintf("nohup%v /home/%s/fablab/bin/%s run --log-formatter pfxlog /home/%s/fablab/cfg/%s > logs/%s.log 2>&1 &",
		sudoCmd, factory.User(), name, factory.User(), cfg, logName)
	if value, err := RemoteExec(factory, serviceCmd); err == nil {
		if len(value) > 0 {
			logrus.Infof("output [%s]", strings.Trim(value, " \t\r\n"))
		}
	} else {
		return err
	}
	return nil
}

func KillService(factory SshConfigFactory, name string) error {
	return RemoteKill(factory, fmt.Sprintf("/home/%s/fablab/bin/%s", factory.User(), name))
}

func RemoteShell(factory SshConfigFactory) error {
	config := factory.Config()

	logrus.Infof("shell for [%s]", factory.Address())

	client, err := ssh.Dial("tcp", factory.Address(), config)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}

	fd := int(os.Stdout.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = session.Close()
		_ = term.Restore(fd, oldState)
	}()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := term.GetSize(fd)
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

func RemoteConsole(factory SshConfigFactory, cmd string) error {
	config := factory.Config()
	logrus.Infof("console for [%s]: '%s'", factory.Address(), cmd)

	client, err := ssh.Dial("tcp", factory.Address(), config)
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

func RemoteExec(sshConfig SshConfigFactory, cmd string) (string, error) {
	return RemoteExecAll(sshConfig, cmd)
}

func RemoteExecAll(sshConfig SshConfigFactory, cmds ...string) (string, error) {
	var b bytes.Buffer
	err := RemoteExecAllTo(sshConfig, &b, cmds...)
	return b.String(), err
}

func RemoteExecAllTo(sshConfig SshConfigFactory, out io.Writer, cmds ...string) error {
	if len(cmds) == 0 {
		return nil
	}
	config := sshConfig.Config()

	logrus.Infof("executing [%s]: '%s'", sshConfig.Address(), cmds[0])

	client, err := ssh.Dial("tcp", sshConfig.Address(), config)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	for idx, cmd := range cmds {
		session, err := client.NewSession()
		if err != nil {
			return err
		}
		session.Stdout = out

		if idx > 0 {
			logrus.Infof("executing [%s]: '%s'", sshConfig.Address(), cmd)
		}
		err = session.Run(cmd)
		_ = session.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func RemoteKill(factory SshConfigFactory, match string) error {
	return RemoteKillFilter(factory, match, "")
}

func RemoteKillFilter(factory SshConfigFactory, match string, anti string) error {
	output, err := RemoteExec(factory, "ps ax")
	if err != nil {
		return errors.Wrapf(err, "unable to get remote process listing [%s]", factory.Address())
	}

	var pidList []int
	r := strings.NewReader(output)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if killMatch(line, match, anti) {
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
		killCmd := "sudo kill"
		for _, pid := range pidList {
			killCmd += fmt.Sprintf(" %d", pid)
		}

		output, err = RemoteExec(factory, killCmd)
		if err != nil {
			return fmt.Errorf("unable to execute [%v] on [%s] (%s). Output: [%v]", killCmd, factory.Address(), err, output)
		}
	}

	return nil
}

func killMatch(s, search, anti string) bool {
	match := false
	if strings.Contains(s, search) {
		match = true
	}
	if anti != "" && strings.Contains(s, anti) {
		match = false
	}
	return match
}

func RemoteFileList(factory SshConfigFactory, path string) ([]os.FileInfo, error) {
	config := factory.Config()

	conn, err := ssh.Dial("tcp", factory.Address(), config)
	if err != nil {
		return nil, fmt.Errorf("error dialing ssh server (%w)", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, fmt.Errorf("error creating sftp client (%w)", err)
	}
	defer func() { _ = client.Close() }()

	files, err := client.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error retrieving directory [%s] (%w)", path, err)
	}

	return files, nil
}

func Chmod(factory SshConfigFactory, remotePath string, mode os.FileMode) error {
	config := factory.Config()

	conn, err := ssh.Dial("tcp", factory.Address(), config)
	if err != nil {
		return errors.Wrap(err, "error dialing ssh server")
	}
	defer func() { _ = conn.Close() }()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return errors.Wrap(err, "error creating sftp client")
	}
	defer func() { _ = client.Close() }()

	rmtFile, err := client.OpenFile(remotePath, os.O_WRONLY)

	if err != nil {
		return errors.Wrapf(err, "unable to open remote file %v", remotePath)
	}
	defer func() { _ = rmtFile.Close() }()

	if err := rmtFile.Chmod(mode); err != nil {
		return errors.Wrapf(err, "unable to chmod remote file %v", remotePath)
	}

	return nil
}

func SendData(factory SshConfigFactory, data []byte, remotePath string) error {
	config := factory.Config()

	conn, err := ssh.Dial("tcp", factory.Address(), config)
	if err != nil {
		return errors.Wrap(err, "error dialing ssh server")
	}
	defer func() { _ = conn.Close() }()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return errors.Wrap(err, "error creating sftp client")
	}
	defer func() { _ = client.Close() }()

	path.Dir(remotePath)
	logrus.Infof("Creating paths %s", path.Dir(remotePath))
	if err := client.MkdirAll(path.Dir(remotePath)); err != nil {
		return errors.Wrapf(err, "unable to create directories for %v", remotePath)
	}

	rmtFile, err := client.OpenFile(remotePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)

	if err != nil {
		return errors.Wrapf(err, "unable to open remote file %v", remotePath)
	}
	defer func() { _ = rmtFile.Close() }()

	_, err = rmtFile.Write(data)

	if err != nil {
		return err
	}

	return nil
}

func SendFile(factory SshConfigFactory, localPath string, remotePath string) error {
	localFile, err := os.ReadFile(localPath)

	if err != nil {
		return errors.Wrapf(err, "unable to read local file %v", localFile)
	}

	return SendData(factory, localFile, remotePath)
}

func RetrieveRemoteFiles(factory SshConfigFactory, localPath string, paths ...string) error {
	if len(paths) < 1 {
		return nil
	}

	if err := os.MkdirAll(localPath, os.ModePerm); err != nil {
		return fmt.Errorf("error creating local path")
	}

	config := factory.Config()

	conn, err := ssh.Dial("tcp", factory.Address(), config)
	if err != nil {
		return fmt.Errorf("error dialing ssh server (%w)", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return fmt.Errorf("error creating sftp client (%w)", err)
	}
	defer func() { _ = client.Close() }()

	for _, path := range paths {
		rf, err := client.Open(path)
		if err != nil {
			return fmt.Errorf("error opening remote file [%s] (%w)", path, err)
		}
		defer func() { _ = rf.Close() }()

		lf, err := os.OpenFile(filepath.Join(localPath, filepath.Base(path)), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error opening local file [%s] (%w)", path, err)
		}
		defer func() { _ = lf.Close() }()

		n, err := io.Copy(lf, rf)
		if err != nil {
			return fmt.Errorf("error copying remote file to local [%s] (%w)", path, err)
		}
		logrus.Infof("%s => %s", path, info.ByteCount(n))
	}

	return nil
}

func DeleteRemoteFiles(factory SshConfigFactory, paths ...string) error {
	config := factory.Config()

	conn, err := ssh.Dial("tcp", factory.Address(), config)
	if err != nil {
		return fmt.Errorf("error dialing ssh server (%w)", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return fmt.Errorf("error creating sftp client (%w)", err)
	}
	defer func() { _ = client.Close() }()

	for _, path := range paths {
		if err := client.Remove(path); err != nil {
			return fmt.Errorf("error removing path [%s] (%w)", path, err)
		}
		logrus.Infof("%s removed", path)
	}

	return nil
}

type SshConfigFactory interface {
	Address() string
	Hostname() string
	Port() int
	User() string
	Config() *ssh.ClientConfig
	KeyPath() string
}

type SshConfigFactoryImpl struct {
	user            string
	host            string
	port            int
	keyPath         string
	resolveAuthOnce sync.Once
	authMethods     []ssh.AuthMethod
}

func NewSshConfigFactory(h *model.Host) *SshConfigFactoryImpl {
	user := h.MustStringVariable("credentials.ssh.username")
	keyPath := h.MustStringVariable("credentials.ssh.key_path")
	factory := &SshConfigFactoryImpl{
		user:    user,
		host:    h.PublicIp,
		port:    22,
		keyPath: keyPath,
	}

	return factory
}

func (factory *SshConfigFactoryImpl) User() string {
	return factory.user
}
func (factory *SshConfigFactoryImpl) Hostname() string {
	return factory.host
}

func (factory *SshConfigFactoryImpl) Port() int {
	return factory.port
}

func (factory *SshConfigFactoryImpl) KeyPath() string {
	return factory.keyPath
}

func (factory *SshConfigFactoryImpl) Address() string {
	return factory.host + ":" + strconv.Itoa(factory.port)
}

func (factory *SshConfigFactoryImpl) Config() *ssh.ClientConfig {
	factory.resolveAuthOnce.Do(func() {
		var methods []ssh.AuthMethod

		if fileMethod, err := sshAuthMethodFromFile(factory.keyPath); err == nil {
			methods = append(methods, fileMethod)
		} else {
			logrus.Error(err)
		}

		if agentMethod := sshAuthMethodAgent(); agentMethod != nil {
			methods = append(methods, sshAuthMethodAgent())
		}

		factory.authMethods = methods
	})

	return &ssh.ClientConfig{
		User:            factory.user,
		Auth:            factory.authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}

func sshAuthMethodFromFile(keyPath string) (ssh.AuthMethod, error) {
	content, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read ssh file [%s]: %w", keyPath, err)
	}

	if signer, err := ssh.ParsePrivateKey(content); err == nil {
		return ssh.PublicKeys(signer), nil
	} else {
		if err.Error() == "ssh: no key found" {
			return nil, fmt.Errorf("no private key found in [%s]: %w", keyPath, err)
		} else if err.(*ssh.PassphraseMissingError) != nil {
			return nil, fmt.Errorf("file is password protected [%s] %w", keyPath, err)
		} else {
			return nil, fmt.Errorf("error parsing private key from [%s]L %w", keyPath, err)
		}
	}
}
