/*
	(c) Copyright NetFoundry Inc. Inc.

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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/openziti/fablab/kernel/libssh"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(newScpCmd())
}

type scpCmd struct {
	forceBuiltIn bool
	recursive    bool
}

func newScpCmd() *cobra.Command {
	cmd := &scpCmd{}

	cobraCmd := &cobra.Command{
		Use:   "scp <src> <dst>",
		Short: "copy files to/from hosts in the model",
		Long: `Copy files between local and remote hosts using scp syntax.
Remote paths use hostSpec:path format, where hostSpec is a fablab selector.

Examples:
  fablab scp ctrl1:./logs/ctrl1.log ./ctrl1.log
  fablab scp ./bin/ziti router-east:./fablab/bin/ziti
  fablab scp ctrl1:./test.file router-east:./test.file
  fablab scp -r ctrl1:./logs/ ./logs/`,
		Args: cobra.ExactArgs(2),
		RunE: cmd.run,
	}

	cobraCmd.Flags().BoolVarP(&cmd.forceBuiltIn, "force-built-in", "f", false,
		"Force use of built-in scp, don't try and detect/use an external scp client")
	cobraCmd.Flags().BoolVarP(&cmd.recursive, "recursive", "r", false,
		"Recursively copy directories")

	return cobraCmd
}

// scpArg represents a parsed scp source or destination argument.
type scpArg struct {
	hostSpec string
	path     string
}

func (a scpArg) isRemote() bool {
	return a.hostSpec != ""
}

// parseScpArg splits a user argument into an optional host specifier and a path.
// Paths starting with /, ./, or ../ are always treated as local.
// Otherwise, the first : is used as a delimiter between hostSpec and path.
func parseScpArg(arg string) scpArg {
	if strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, "./") || strings.HasPrefix(arg, "../") {
		return scpArg{path: arg}
	}
	if hostSpec, filePath, ok := strings.Cut(arg, ":"); ok {
		return scpArg{hostSpec: hostSpec, path: filePath}
	}
	return scpArg{path: arg}
}

func (self *scpCmd) run(_ *cobra.Command, args []string) error {
	if err := model.Bootstrap(); err != nil {
		return fmt.Errorf("unable to bootstrap (%w)", err)
	}

	m := model.GetModel()

	src := parseScpArg(args[0])
	dst := parseScpArg(args[1])

	if !src.isRemote() && !dst.isRemote() {
		return fmt.Errorf("at least one of source or destination must be remote (use hostSpec:path syntax)")
	}

	// Resolve source host (must be exactly 1 if remote)
	var srcHost *model.Host
	if src.isRemote() {
		host, err := m.SelectHost(src.hostSpec)
		if err != nil {
			return fmt.Errorf("source host: %w", err)
		}
		srcHost = host
	}

	// Resolve destination hosts (can be multiple if remote)
	var dstHosts []*model.Host
	if dst.isRemote() {
		dstHosts = m.SelectHosts(dst.hostSpec)
		if len(dstHosts) == 0 {
			return fmt.Errorf("destination selector [%s] matched 0 hosts", dst.hostSpec)
		}
	}

	if !self.forceBuiltIn {
		if _, err := exec.LookPath("scp"); err == nil {
			return self.nativeScp(srcHost, src.path, dstHosts, dst.path)
		}
	}

	return self.builtinScp(srcHost, src.path, dstHosts, dst.path)
}

func (self *scpCmd) nativeScp(srcHost *model.Host, srcPath string, dstHosts []*model.Host, dstPath string) error {
	// Determine which host to use for SSH config (key, port)
	var refHost *model.Host
	if srcHost != nil {
		refHost = srcHost
	} else {
		refHost = dstHosts[0]
	}
	sshCfg := refHost.NewSshConfigFactory()

	formatRemotePath := func(host *model.Host, filePath string) string {
		cfg := host.NewSshConfigFactory()
		return cfg.User() + "@" + cfg.Hostname() + ":" + filePath
	}

	// Build the source argument
	nativeSrc := srcPath
	if srcHost != nil {
		nativeSrc = formatRemotePath(srcHost, srcPath)
	}

	// If destination is local, run a single scp
	if len(dstHosts) == 0 {
		return self.runNativeScp(sshCfg, nativeSrc, dstPath)
	}

	// One scp per destination host
	var lastErr error
	for _, dstHost := range dstHosts {
		nativeDst := formatRemotePath(dstHost, dstPath)
		logrus.Infof("scp %s -> %s", nativeSrc, nativeDst)
		if err := self.runNativeScp(sshCfg, nativeSrc, nativeDst); err != nil {
			logrus.Errorf("scp to %s failed: %v", dstHost.PublicIp, err)
			lastErr = err
		}
	}
	return lastErr
}

func (self *scpCmd) runNativeScp(sshCfg *libssh.SshConfigFactoryImpl, src, dst string) error {
	cmdArgs := []string{
		"-i", sshCfg.KeyPath(),
		"-o", "StrictHostKeyChecking=no",
	}
	if self.recursive {
		cmdArgs = append(cmdArgs, "-r")
	}
	if sshCfg.Port() != 22 {
		cmdArgs = append(cmdArgs, "-P", fmt.Sprintf("%d", sshCfg.Port()))
	}
	cmdArgs = append(cmdArgs, src, dst)

	cmd := exec.Command("scp", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (self *scpCmd) builtinScp(srcHost *model.Host, srcPath string, dstHosts []*model.Host, dstPath string) error {
	switch {
	case srcHost == nil:
		// Local to remote
		return self.builtinUpload(srcPath, dstHosts, dstPath)
	case len(dstHosts) == 0:
		// Remote to local
		return self.builtinDownload(srcHost, srcPath, dstPath)
	default:
		// Remote to remote: download to temp, then upload to each dest
		return self.builtinRemoteToRemote(srcHost, srcPath, dstHosts, dstPath)
	}
}

func (self *scpCmd) builtinUpload(srcPath string, dstHosts []*model.Host, dstPath string) error {
	// Expand local wildcards
	matches, err := filepath.Glob(srcPath)
	if err != nil {
		return fmt.Errorf("invalid glob pattern [%s]: %w", srcPath, err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("no files matched [%s]", srcPath)
	}

	var lastErr error
	for _, dstHost := range dstHosts {
		sshCfg := dstHost.NewSshConfigFactory()
		for _, match := range matches {
			fi, err := os.Stat(match)
			if err != nil {
				return fmt.Errorf("unable to stat [%s]: %w", match, err)
			}
			if fi.IsDir() {
				if !self.recursive {
					return fmt.Errorf("[%s] is a directory; use -r for recursive copy", match)
				}
				logrus.Infof("uploading directory %s -> %s:%s", match, dstHost.PublicIp, dstPath)
				if err := libssh.SendDirectory(sshCfg, match, dstPath); err != nil {
					logrus.Errorf("upload to %s failed: %v", dstHost.PublicIp, err)
					lastErr = err
				}
			} else {
				remoteDst := dstPath
				// If multiple matches or dstPath looks like a directory, append the filename
				if len(matches) > 1 || strings.HasSuffix(dstPath, "/") {
					remoteDst = dstPath + "/" + filepath.Base(match)
				}
				logrus.Infof("uploading %s -> %s:%s", match, dstHost.PublicIp, remoteDst)
				if err := libssh.SendFile(sshCfg, match, remoteDst); err != nil {
					logrus.Errorf("upload to %s failed: %v", dstHost.PublicIp, err)
					lastErr = err
				}
			}
		}
	}
	return lastErr
}

func (self *scpCmd) builtinDownload(srcHost *model.Host, srcPath string, dstPath string) error {
	sshCfg := srcHost.NewSshConfigFactory()

	// Expand remote wildcards if the path contains glob characters
	paths := []string{srcPath}
	if strings.ContainsAny(srcPath, "*?[") {
		expanded, err := libssh.RemoteGlob(sshCfg, srcPath)
		if err != nil {
			return fmt.Errorf("remote glob failed: %w", err)
		}
		if len(expanded) == 0 {
			return fmt.Errorf("no remote files matched [%s]", srcPath)
		}
		paths = expanded
	}

	return libssh.RetrieveRemoteFiles(sshCfg, dstPath, paths...)
}

func (self *scpCmd) builtinRemoteToRemote(srcHost *model.Host, srcPath string, dstHosts []*model.Host, dstPath string) error {
	tmpDir, err := os.MkdirTemp("", "fablab-scp-*")
	if err != nil {
		return fmt.Errorf("unable to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	logrus.Infof("downloading from %s to staging directory", srcHost.PublicIp)
	if err := self.builtinDownload(srcHost, srcPath, tmpDir); err != nil {
		return fmt.Errorf("download from source failed: %w", err)
	}

	// Upload each file/dir from the temp directory to destinations
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("unable to read staging directory: %w", err)
	}

	var lastErr error
	for _, dstHost := range dstHosts {
		sshCfg := dstHost.NewSshConfigFactory()
		for _, entry := range entries {
			localPath := filepath.Join(tmpDir, entry.Name())
			remoteDst := dstPath
			if len(entries) > 1 || strings.HasSuffix(dstPath, "/") {
				remoteDst = dstPath + "/" + entry.Name()
			}
			if entry.IsDir() {
				logrus.Infof("uploading directory %s -> %s:%s", entry.Name(), dstHost.PublicIp, remoteDst)
				if err := libssh.SendDirectory(sshCfg, localPath, remoteDst); err != nil {
					logrus.Errorf("upload to %s failed: %v", dstHost.PublicIp, err)
					lastErr = err
				}
			} else {
				logrus.Infof("uploading %s -> %s:%s", entry.Name(), dstHost.PublicIp, remoteDst)
				if err := libssh.SendFile(sshCfg, localPath, remoteDst); err != nil {
					logrus.Errorf("upload to %s failed: %v", dstHost.PublicIp, err)
					lastErr = err
				}
			}
		}
	}
	return lastErr
}
