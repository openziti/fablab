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

package rsync

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/libssh"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	defaultSyncRetries      = 3
	defaultSyncRetryBackoff = 5 * time.Second
)

func RsyncStaged() model.Stage {
	return &stagedRsyncStage{
		hostSelector: "*",
	}
}

func RsyncSelected(hosts, src, dst string) model.Stage {
	return &stagedRsyncStage{
		hostSelector: hosts,
		src:          src,
		dst:          dst,
	}
}

type stagedRsyncStage struct {
	hostSelector string
	src          string
	dst          string
}

// rsync to first host
// rsync from first host to next host in region

func (rsync *stagedRsyncStage) Execute(run model.Run) error {
	group, ctx := errgroup.WithContext(context.Background())
	hosts := map[string]*model.Host{}

	for _, host := range run.GetModel().SelectHosts(rsync.hostSelector) {
		hosts[host.GetPath()] = host
	}

	localSyncer := &localRsyncer{
		rsyncContext: &rsyncContext{
			hosts: hosts,
			group: group,
			ctx:   ctx,
			src:   rsync.src,
			dst:   rsync.dst,
		},
		regions: map[string]struct{}{},
	}

	localSyncer.init(run.GetModel())

	pfxlog.Logger().Infof("rsyncing %d hosts %s -> %s",
		len(hosts), localSyncer.src, localSyncer.dst)

	group.Go(localSyncer.run)

	return group.Wait()
}

type rsyncContext struct {
	hosts map[string]*model.Host
	sync.Mutex
	group   *errgroup.Group
	ctx     context.Context
	syncing int
	src     string
	dst     string
}

func (self *rsyncContext) init(m *model.Model) {
	if self.src == "" {
		syncTarget := m.GetStringVariableOr("sync.target", "all")
		extraPath := ""

		self.src = model.KitBuild()
		if syncTarget != "all" {
			extraPath = syncTarget + "/"
			self.src = filepath.Join(self.src, syncTarget)
		}
		self.dst = "fablab/" + extraPath
	}

	if self.src != "" && !strings.HasSuffix(self.src, "/") {
		self.src += "/"
	}

	if self.dst != "" && !strings.HasSuffix(self.dst, "/") {
		self.dst += "/"
	}
}

func (self *rsyncContext) getDestPath(h *model.Host) string {
	if !strings.HasPrefix(self.dst, "/") {
		return fmt.Sprintf("/home/%s/%s", h.GetSshUser(), self.dst)
	}
	return self.dst
}

func (self *rsyncContext) GetNextHostPreferringRegion(regionId string) (*model.Host, int, int) {
	self.Lock()
	defer self.Unlock()

	var next *model.Host
	for _, host := range self.hosts {
		next = host
		if host.Region.Id == regionId {
			break
		}
	}
	if next != nil {
		delete(self.hosts, next.GetPath())
		self.syncing++
	}
	return next, len(self.hosts), self.syncing
}

func (self *rsyncContext) GetNextHostPreferringNotRegions(regionIds map[string]struct{}) (*model.Host, int, int) {
	self.Lock()
	defer self.Unlock()

	var next *model.Host
	for _, host := range self.hosts {
		next = host
		if _, found := regionIds[host.Region.Id]; !found {
			break
		}
	}
	if next != nil {
		delete(self.hosts, next.GetPath())
		self.syncing++
	}
	return next, len(self.hosts), self.syncing
}

func (self *rsyncContext) markDone() (int, int) {
	self.Lock()
	defer self.Unlock()
	self.syncing--
	return len(self.hosts), self.syncing
}

type localRsyncer struct {
	*rsyncContext
	regions map[string]struct{}
}

func (self *localRsyncer) run() error {
	for {
		host, left, current := self.GetNextHostPreferringNotRegions(self.regions)
		if host == nil {
			return nil
		}
		logrus.Infof("syncing local -> %v. Left: %v, current: %v", host.PublicIp, left, current)
		config := NewConfig(host)
		if err := synchronizeHost(self.rsyncContext, config); err != nil {
			return errors.Wrapf(err, "error synchronizing host [%s/%s]", host.GetRegion().GetId(), host.GetId())
		}
		left, current = self.markDone()
		logrus.Infof("finished syncing local -> %v. Left: %v, current: %v", host.PublicIp, left, current)

		if err := self.ctx.Err(); err != nil {
			logrus.WithError(err).Info("exiting sync early as group context is failed")
			return err
		}

		remoteSyncer := &remoteRsyncer{
			host:         host,
			rsyncContext: self.rsyncContext,
		}

		self.group.Go(remoteSyncer.run)
	}
}

type remoteRsyncer struct {
	host *model.Host
	*rsyncContext
}

func (self *remoteRsyncer) run() error {
	didAltRegion := false
	for {
		// only do one remote region, otherwise let remote region handle itself
		if didAltRegion {
			return nil
		}
		host, left, current := self.GetNextHostPreferringRegion(self.host.Region.Id)
		if host == nil {
			return nil
		}
		logrus.Infof("syncing %v -> %v. Left: %v, current: %v", self.host.PublicIp, host.PublicIp, left, current)
		if host.Region.Id != self.host.Region.Id {
			didAltRegion = true
		}
		srcConfig := NewConfig(self.host)
		dstConfig := NewConfig(host)
		if err := synchronizeHostToHost(self.rsyncContext, srcConfig, dstConfig); err != nil {
			return errors.Wrapf(err, "error synchronizing host [%s/%s]", host.GetRegion().GetId(), host.GetId())
		}
		left, current = self.markDone()
		logrus.Infof("finished syncing %v -> %v. Left: %v, current: %v", self.host.PublicIp, host.PublicIp, left, current)

		if err := self.ctx.Err(); err != nil {
			logrus.WithError(err).Info("exiting sync early as group context is failed")
			return err
		}

		remoteSyncer := &remoteRsyncer{
			host:         host,
			rsyncContext: self.rsyncContext,
		}

		self.group.Go(remoteSyncer.run)
	}
}

func synchronizeHost(ctx *rsyncContext, config *Config) error {
	var lastErr error
	for attempt := 1; attempt <= defaultSyncRetries; attempt++ {
		lastErr = synchronizeHostOnce(ctx, config)
		if lastErr == nil {
			return nil
		}
		if attempt < defaultSyncRetries {
			logrus.WithError(lastErr).Warnf("rsync to %s failed (attempt %d/%d), retrying in %v",
				config.host.PublicIp, attempt, defaultSyncRetries, defaultSyncRetryBackoff)
			time.Sleep(defaultSyncRetryBackoff)
		}
	}
	return lastErr
}

func synchronizeHostOnce(ctx *rsyncContext, config *Config) error {
	mkdirCmd := fmt.Sprintf("mkdir -p %s", ctx.dst)
	if output, err := libssh.RemoteExec(config.sshConfigFactory, mkdirCmd); err == nil {
		if output != "" {
			logrus.Infof("output [%s]", strings.Trim(output, " \t\r\n"))
		}
	} else {
		return err
	}

	destination := fmt.Sprintf("%s:%s", config.loginPrefix(), ctx.getDestPath(config.host))
	if err := RunRsync(config, ctx.src, destination); err != nil {
		return fmt.Errorf("rsyncStage failed (%w)", err)
	}

	return nil
}

func synchronizeHostToHost(ctx *rsyncContext, srcConfig, dstConfig *Config) error {
	var lastErr error
	for attempt := 1; attempt <= defaultSyncRetries; attempt++ {
		lastErr = synchronizeHostToHostOnce(ctx, srcConfig, dstConfig)
		if lastErr == nil {
			return nil
		}
		if attempt < defaultSyncRetries {
			logrus.WithError(lastErr).Warnf("rsync %s -> %s failed (attempt %d/%d), retrying in %v",
				srcConfig.host.PublicIp, dstConfig.host.PublicIp, attempt, defaultSyncRetries, defaultSyncRetryBackoff)
			time.Sleep(defaultSyncRetryBackoff)
		}
	}
	return lastErr
}

func synchronizeHostToHostOnce(ctx *rsyncContext, srcConfig, dstConfig *Config) error {
	mkdirCmd := fmt.Sprintf("mkdir -p %s", ctx.dst)
	if output, err := libssh.RemoteExec(dstConfig.sshConfigFactory, mkdirCmd); err == nil {
		if output != "" {
			logrus.Infof("output [%s]", strings.Trim(output, " \t\r\n"))
		}
	} else {
		return err
	}

	destination := fmt.Sprintf("%s:%s", dstConfig.loginPrefix(), ctx.getDestPath(dstConfig.host))

	cmd := fmt.Sprintf("rsync -avz --delete -e 'ssh -o StrictHostKeyChecking=no' %s %s",
		ctx.getDestPath(srcConfig.host), destination)
	output, err := libssh.RemoteExec(srcConfig.sshConfigFactory, cmd)
	if err == nil && output != "" {
		logrus.Infof("output [%s]", strings.Trim(output, " \t\r\n"))
	}
	return err
}

type Config struct {
	host             *model.Host
	sshBin           string
	sshConfigFactory libssh.SshConfigFactory
	rsyncBin         string
}

func NewConfig(h *model.Host) *Config {
	config := &Config{
		host:             h,
		sshBin:           h.GetStringVariableOr("distribution.ssh_bin", "ssh"),
		sshConfigFactory: h.NewSshConfigFactory(),
		rsyncBin:         h.GetStringVariableOr("distribution.rsync_bin", "rsync"),
	}

	return config
}

func (config *Config) sshIdentityFlag() string {
	if config.sshConfigFactory.KeyPath() != "" {
		return "-i " + config.sshConfigFactory.KeyPath()
	}

	return ""
}

func (config *Config) loginPrefix() string {
	return config.sshConfigFactory.User() + "@" + config.sshConfigFactory.Hostname()
}

func (config *Config) SshCommand() string {
	return config.sshBin + " " + config.sshIdentityFlag()
}

func NewRsyncHost(hostSpec, src, dest string) model.Stage {
	return &rsyncHostStage{
		hostSpec: hostSpec,
		src:      src,
		dest:     dest,
	}
}

type rsyncHostStage struct {
	hostSpec string
	src      string
	dest     string
}

func (self *rsyncHostStage) Execute(run model.Run) error {
	return run.GetModel().ForEachHost(self.hostSpec, 1, func(host *model.Host) error {
		cfg := NewConfig(host)
		dest := cfg.sshConfigFactory.User() + "@" + cfg.sshConfigFactory.Hostname() + ":" + self.dest
		return RunRsync(NewConfig(host), self.src, dest)
	})
}
