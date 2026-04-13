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
	"runtime"
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
	recursive bool
}

func newScpCmd() *cobra.Command {
	cmd := &scpCmd{}

	cobraCmd := &cobra.Command{
		Use:   "scp <src> <dst>",
		Short: "copy files to/from hosts in the model",
		Long: `Copy files between local and remote hosts using scp syntax.
Remote paths use hostSpec:path format, where hostSpec is a fablab selector.
Delegates to the native scp client, which must be available on PATH.

Examples:
  fablab scp ctrl1:./logs/ctrl1.log ./ctrl1.log
  fablab scp ./bin/ziti router-east:./fablab/bin/ziti
  fablab scp ctrl1:./test.file router-east:./test.file
  fablab scp -r ctrl1:./logs/ ./logs/`,
		Args: cobra.ExactArgs(2),
		RunE: cmd.run,
	}

	cobraCmd.Flags().BoolVarP(&cmd.recursive, "recursive", "r", false,
		"Recursively copy directories")

	return cobraCmd
}

// nativeCanReachAll reports whether a remote-to-remote scp can serve every destination
// host from the source host. Native scp applies a single key and port to both endpoints,
// so this holds only when all destinations share the source's key path and port.
func nativeCanReachAll(srcHost *model.Host, dstHosts []*model.Host) bool {
	srcCfg := srcHost.NewSshConfigFactory()
	for _, dstHost := range dstHosts {
		dstCfg := dstHost.NewSshConfigFactory()
		if srcCfg.KeyPath() != dstCfg.KeyPath() || srcCfg.Port() != dstCfg.Port() {
			return false
		}
	}
	return true
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
	// On Windows, a volume-qualified path is local; don't mistake its drive letter for a
	// remote host selector. VolumeName covers drive-absolute (C:\work), drive-relative
	// (C:work), extended-length (\\?\C:\...), and UNC (\\server\share) forms. It returns
	// "" on non-Windows builds, so Unix parsing is unaffected.
	if runtime.GOOS == "windows" && filepath.VolumeName(arg) != "" {
		return scpArg{path: arg}
	}
	if hostSpec, filePath, ok := strings.Cut(arg, ":"); ok {
		// A bare "host:" refers to the remote home directory, matching native scp.
		if filePath == "" {
			filePath = "."
		}
		return scpArg{hostSpec: hostSpec, path: filePath}
	}
	return scpArg{path: arg}
}

func (self *scpCmd) run(_ *cobra.Command, args []string) error {
	if err := model.Bootstrap(); err != nil {
		return fmt.Errorf("unable to bootstrap (%w)", err)
	}

	if _, err := exec.LookPath("scp"); err != nil {
		return fmt.Errorf("scp client not found on PATH; fablab scp delegates to a native scp binary (%w)", err)
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

	// Native scp applies a single -i key and -P port to both endpoints (via -3 for a
	// remote-to-remote copy, which routes through the local host where the key lives),
	// so remote-to-remote only works when the two remotes share a key and port.
	if srcHost != nil && len(dstHosts) > 0 && !nativeCanReachAll(srcHost, dstHosts) {
		return fmt.Errorf("remote-to-remote scp requires the source and destination hosts to share an ssh key and port")
	}

	return self.nativeScp(srcHost, src.path, dstHosts, dst.path)
}

func (self *scpCmd) nativeScp(srcHost *model.Host, srcPath string, dstHosts []*model.Host, dstPath string) error {
	formatRemotePath := func(host *model.Host, filePath string) string {
		cfg := host.NewSshConfigFactory()
		return cfg.User() + "@" + cfg.Hostname() + ":" + filePath
	}

	// Build the source arguments. A remote source is passed as-is (the remote shell
	// expands any globs); a local source may be a glob, which native scp does not expand
	// itself, so expand it here.
	var nativeSrcs []string
	if srcHost != nil {
		nativeSrcs = []string{formatRemotePath(srcHost, srcPath)}
	} else {
		matches, err := filepath.Glob(srcPath)
		if err != nil {
			return fmt.Errorf("invalid glob pattern [%s]: %w", srcPath, err)
		}
		if len(matches) == 0 {
			// Literal path, or no match; let scp report the error against the original.
			matches = []string{srcPath}
		}
		nativeSrcs = matches
	}

	// Remote source, local destination: use the source host's ssh config.
	if len(dstHosts) == 0 {
		return self.runNativeScp(srcHost.NewSshConfigFactory(), dstPath, false, nativeSrcs...)
	}

	// A remote source with remote destinations must be staged through the local host.
	remoteToRemote := srcHost != nil

	// One scp per destination host.
	var lastErr error
	for _, dstHost := range dstHosts {
		// For a local source, each destination is reached with its own ssh config. For a
		// remote-to-remote copy, -3 relays through the local host with a single key/port
		// that all endpoints share (verified before dispatch), so use the source's config.
		sshCfg := dstHost.NewSshConfigFactory()
		if remoteToRemote {
			sshCfg = srcHost.NewSshConfigFactory()
		}

		nativeDst := formatRemotePath(dstHost, dstPath)
		logrus.Infof("scp %v -> %s", nativeSrcs, nativeDst)
		if err := self.runNativeScp(sshCfg, nativeDst, remoteToRemote, nativeSrcs...); err != nil {
			logrus.Errorf("scp to %s failed: %v", dstHost.PublicIp, err)
			lastErr = err
		}
	}
	return lastErr
}

func (self *scpCmd) runNativeScp(sshCfg *libssh.SshConfigFactoryImpl, dst string, viaLocal bool, srcs ...string) error {
	cmdArgs := []string{
		"-i", sshCfg.KeyPath(),
		"-o", "StrictHostKeyChecking=no",
	}
	if viaLocal {
		// -3 routes a remote-to-remote copy through the local host, which is where the
		// ssh key lives; without it scp would try to connect directly between the two
		// remotes, which have no credentials for each other.
		cmdArgs = append(cmdArgs, "-3")
	}
	if self.recursive {
		cmdArgs = append(cmdArgs, "-r")
	}
	if sshCfg.Port() != 22 {
		cmdArgs = append(cmdArgs, "-P", fmt.Sprintf("%d", sshCfg.Port()))
	}
	cmdArgs = append(cmdArgs, srcs...)
	cmdArgs = append(cmdArgs, dst)

	cmd := exec.Command("scp", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
