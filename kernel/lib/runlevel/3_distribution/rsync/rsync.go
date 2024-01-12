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

package rsync

import (
	"context"
	"fmt"
	"github.com/openziti/fablab/kernel/libssh"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"strings"
	"sync"
)

func Rsync(concurrency int) model.Stage {
	return &rsyncStage{
		concurrency: concurrency,
	}
}

func (rsync *rsyncStage) Execute(run model.Run) error {
	return run.GetModel().ForEachHost("*", rsync.concurrency, func(host *model.Host) error {
		config := NewConfig(host)
		if err := synchronizeHost(config); err != nil {
			return fmt.Errorf("error synchronizing host [%s/%s] (%s)", host.GetRegion().GetId(), host.GetId(), err)
		}
		return nil
	})
}

type rsyncStage struct {
	concurrency int
}

func RsyncStaged() model.Stage {
	return &stagedRsyncStage{}
}

type stagedRsyncStage struct {
}

// rsync to first host
// rsync from first host to next host in region

func (rsync *stagedRsyncStage) Execute(run model.Run) error {
	group, ctx := errgroup.WithContext(context.Background())
	hosts := map[string]*model.Host{}

	run.GetModel().RangeSortedRegions(func(id string, region *model.Region) {
		region.RangeSortedHosts(func(id string, host *model.Host) {
			hosts[host.GetPath()] = host
		})
	})

	localSyncer := &localRsyncer{
		rsyncContext: &rsyncContext{
			hosts: hosts,
			group: group,
			ctx:   ctx,
		},
		regions: map[string]struct{}{},
	}

	group.Go(localSyncer.run)

	return group.Wait()
}

type rsyncContext struct {
	hosts map[string]*model.Host
	sync.Mutex
	group   *errgroup.Group
	ctx     context.Context
	syncing int
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
		if err := synchronizeHost(config); err != nil {
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
		host, left, current := self.rsyncContext.GetNextHostPreferringRegion(self.host.Region.Id)
		if host == nil {
			return nil
		}
		logrus.Infof("syncing %v -> %v. Left: %v, current: %v", self.host.PublicIp, host.PublicIp, left, current)
		if host.Region.Id != self.host.Region.Id {
			didAltRegion = true
		}
		srcConfig := NewConfig(self.host)
		dstConfig := NewConfig(host)
		if err := synchronizeHostToHost(srcConfig, dstConfig); err != nil {
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

func synchronizeHost(config *Config) error {
	if output, err := libssh.RemoteExec(config.sshConfigFactory, "mkdir -p /home/ubuntu/fablab/bin"); err == nil {
		if output != "" {
			logrus.Infof("output [%s]", strings.Trim(output, " \t\r\n"))
		}
	} else {
		return err
	}

	extraPath := config.syncTarget

	if err := RunRsync(config, model.KitBuild()+"/"+extraPath, fmt.Sprintf("ubuntu@%s:/home/ubuntu/fablab/"+extraPath, config.sshConfigFactory.Hostname())); err != nil {
		return fmt.Errorf("rsyncStage failed (%w)", err)
	}

	return nil
}

func synchronizeHostToHost(srcConfig, dstConfig *Config) error {
	if output, err := libssh.RemoteExec(dstConfig.sshConfigFactory, "mkdir -p /home/ubuntu/fablab/bin"); err == nil {
		if output != "" {
			logrus.Infof("output [%s]", strings.Trim(output, " \t\r\n"))
		}
	} else {
		return err
	}

	extraPath := dstConfig.syncTarget

	dst := fmt.Sprintf("ubuntu@%s:/home/ubuntu/fablab/%s", dstConfig.sshConfigFactory.Hostname(), extraPath)
	cmd := fmt.Sprintf("rsync -avz --delete -e 'ssh -o StrictHostKeyChecking=no' /home/ubuntu/fablab/%v* %v", extraPath, dst)
	output, err := libssh.RemoteExec(srcConfig.sshConfigFactory, cmd)
	if err == nil && output != "" {
		logrus.Infof("output [%s]", strings.Trim(output, " \t\r\n"))
	}
	return err
}

type Config struct {
	sshBin           string
	sshConfigFactory libssh.SshConfigFactory
	rsyncBin         string
	syncTarget       string
}

func NewConfig(h *model.Host) *Config {
	config := &Config{
		sshBin:           h.GetStringVariableOr("distribution.ssh_bin", "ssh"),
		sshConfigFactory: h.NewSshConfigFactory(),
		rsyncBin:         h.GetStringVariableOr("distribution.rsync_bin", "rsync"),
	}

	config.syncTarget = h.GetStringVariableOr("sync.target", "all")
	if config.syncTarget == "all" {
		config.syncTarget = ""
	}

	if config.syncTarget != "" && !strings.HasSuffix(config.syncTarget, "/") {
		config.syncTarget += "/"
	}

	return config
}

func (config *Config) sshIdentityFlag() string {
	if config.sshConfigFactory.KeyPath() != "" {
		return "-i " + config.sshConfigFactory.KeyPath()
	}

	return ""
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
