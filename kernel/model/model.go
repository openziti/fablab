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

package model

import (
	"embed"
	"fmt"
	"github.com/openziti/fablab/kernel/lib/figlet"
	"github.com/openziti/fablab/kernel/libssh"
	"github.com/openziti/foundation/v2/info"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	EntityTypeModel              = "model"
	EntityTypeRegion             = "region"
	EntityTypeHost               = "host"
	EntityTypeComponent          = "component"
	EntityTypeParent             = "parent"
	EntityTypeSelfOrParent       = "selfOrParent"
	EntityTypeSelfOrParentSymbol = "^"
	EntityTypeChild              = "child"
	EntityTypeSelfOrChild        = "selfOrChild"
	EntityTypeAny                = "*"
)

func getTraversals(entityType string) (bool, bool, bool) {
	if EntityTypeParent == entityType {
		return true, false, false
	}
	if EntityTypeSelfOrParent == entityType || EntityTypeSelfOrParentSymbol == entityType {
		return true, true, false
	}
	if EntityTypeChild == entityType {
		return false, false, true
	}
	if EntityTypeSelfOrChild == entityType {
		return false, true, true
	}
	if EntityTypeAny == entityType {
		return true, true, true
	}
	return false, false, false
}

func matchHierarchical(entityType string, matcher EntityMatcher, entity Entity) bool {
	checkParent, checkSelf, checkChildren := getTraversals(entityType)

	if checkSelf && matcher(entity) {
		return true
	}

	if parent := entity.GetParentEntity(); parent != nil && checkParent {
		if parent.Matches(EntityTypeSelfOrParent, matcher) {
			return true
		}
	}

	if checkChildren {
		for _, child := range entity.GetChildren() {
			if child.Matches(EntityTypeSelfOrChild, matcher) {
				return true
			}
		}
	}

	return false
}

type EntityVisitor func(Entity)

type Entity interface {
	GetModel() *Model
	GetType() string
	GetId() string
	GetScope() *Scope
	GetParentEntity() Entity
	Accept(EntityVisitor)
	GetChildren() []Entity
	Matches(entityType string, matcher EntityMatcher) bool

	GetVariable(name string) (interface{}, bool)
	GetVariableOr(name string, defaultValue interface{}) interface{}
	MustVariable(name string) interface{}
}

type VariableNamePrefixMapper func(entityPath []string, name string) string
type VariableNameMapper func(string) string
type VariableNameParser func(string) []string

type VarConfig struct {
	VariableNameParser            VariableNameParser
	CommandLineVariableNameMapper VariableNameMapper
	CommandLinePrefixes           []string
	EnvVariableNameMapper         VariableNameMapper
	DefaultVariableResolver       VariableResolver
	DefaultScopedVariableResolver VariableResolver
	SecretsKeys                   []string
	VariableNamePrefixMapper      VariableNamePrefixMapper
	ResolverLogger                func(resolver string, entity Entity, name string, result interface{}, found bool, msgAndArgs ...interface{})
	BindingResolver               *MapVariableResolver
	LabelResolver                 *MapVariableResolver
}

func (self *VarConfig) SetDefaults() {
	if self.VariableNameParser == nil {
		self.VariableNameParser = func(name string) []string {
			return strings.Split(name, ".")
		}
	}

	if self.CommandLineVariableNameMapper == nil {
		self.CommandLineVariableNameMapper = func(s string) string {
			return s
		}
	}

	if len(self.CommandLinePrefixes) == 0 {
		self.CommandLinePrefixes = []string{"-V"}
	}

	if self.EnvVariableNameMapper == nil {
		self.EnvVariableNameMapper = func(s string) string {
			return strings.ToUpper(strings.ReplaceAll(s, ".", "_"))
		}
	}

	self.BindingResolver = NewMapVariableResolver("bindings", bindings)
	self.LabelResolver = NewMapVariableResolver("label", nil)

	if self.DefaultVariableResolver == nil {
		defaultResolverSet := &ChainedVariableResolver{}
		defaultResolverSet.AppendResolver(CmdLineArgVariableResolver{})
		defaultResolverSet.AppendResolver(EnvVariableResolver{})
		defaultResolverSet.AppendResolver(self.LabelResolver)
		defaultResolverSet.AppendResolver(self.BindingResolver)
		defaultResolverSet.AppendResolver(HierarchicalVariableResolver{})
		self.DefaultVariableResolver = defaultResolverSet

		combinedResolvers := &ChainedVariableResolver{}
		combinedResolvers.AppendResolver(NewScopedVariableResolver(defaultResolverSet))
		combinedResolvers.AppendResolver(defaultResolverSet)

		self.DefaultScopedVariableResolver = combinedResolvers
	}

	if len(self.SecretsKeys) == 0 {
		self.SecretsKeys = []string{
			"key", "keys",
			"credential", "credentials",
			"password", "passwords",
			"secret", "secrets",
		}
	}

	if self.VariableNamePrefixMapper == nil {
		self.VariableNamePrefixMapper = func(entityPath []string, name string) string {
			return strings.Join(append(entityPath, name), ".")
		}
	}

	if self.ResolverLogger == nil {
		self.ResolverLogger = func(resolver string, entity Entity, name string, result interface{}, found bool, msgAndArgs ...interface{}) {
		}
	}
}

func (self *VarConfig) EnableDebugLogger() {
	self.ResolverLogger = func(resolver string, entity Entity, name string, result interface{}, found bool, msgAndArgs ...interface{}) {
		msg := ""
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(", ctx=%v", msgAndArgs[0])
			if len(msg) > 1 {
				msg = fmt.Sprintf(msg, msgAndArgs[1:]...)
			}
		}
		fmt.Printf("%v: %v[id=%v] key=%v result=%v, found=%v%v\n", resolver, entity.GetType(), entity.GetId(), name, result, found, msg)
	}
}

type Resource fs.FS

type Resources map[string]Resource

type Model struct {
	Id string

	Scope
	VarConfig           VarConfig
	Regions             Regions
	StructureFactories  []Factory // Factories that change the model structure, eg: add/remove hosts
	Factories           []Factory
	BootstrapExtensions []BootstrapExtension
	Actions             map[string]ActionBinder
	Infrastructure      Stages
	Configuration       Stages
	Distribution        Stages
	Activation          Stages
	Operation           Stages
	Disposal            Stages
	MetricsHandlers     []MetricsHandler
	Resources           Resources

	actions map[string]Action

	initialized atomic.Bool

	regionIds    IdPool
	hostIds      IdPool
	componentIds IdPool
}

func (m *Model) GetModel() *Model {
	return m
}

func (m *Model) GetId() string {
	return m.Id
}

func (m *Model) GetType() string {
	return EntityTypeModel
}

func (m *Model) GetScope() *Scope {
	return &m.Scope
}

func (m *Model) GetParentEntity() Entity {
	return nil
}

func (m *Model) GetResource(name string) fs.FS {
	if resource, found := m.Resources[name]; found {
		return resource
	}
	return embed.FS{}
}

func (m *Model) GetNextRegionIndex() uint32 {
	return m.regionIds.GetNextId()
}

func (m *Model) GetNextHostIndex() uint32 {
	return m.hostIds.GetNextId()
}

func (m *Model) GetNextComponentIndex() uint32 {
	return m.componentIds.GetNextId()
}

func (m *Model) Matches(entityType string, matcher EntityMatcher) bool {
	if EntityTypeModel == entityType {
		return matcher(m)
	}

	if EntityTypeRegion == entityType || EntityTypeHost == entityType || EntityTypeComponent == entityType {
		for _, child := range m.GetChildren() {
			if child.Matches(entityType, matcher) {
				return true
			}
		}
	}

	return matchHierarchical(entityType, matcher, m)
}

func (m *Model) GetChildren() []Entity {
	if len(m.Regions) == 0 {
		return nil
	}

	result := make([]Entity, 0, len(m.Regions))
	for _, entity := range m.Regions {
		result = append(result, entity)
	}
	return result
}

func (m *Model) init() {
	if m.initialized.CompareAndSwap(false, true) {
		m.VarConfig.SetDefaults()

		if m.Data == nil {
			m.Data = Data{}
		}
		m.initialize(m, false)
	}
	m.RangeSortedRegions(func(id string, region *Region) {
		region.init(id, m)
	})
}

func (m *Model) Accept(visitor EntityVisitor) {
	visitor(m)
	for _, region := range m.Regions {
		region.Accept(visitor)
	}
}

func (m *Model) RemoveRegion(region *Region) {
	delete(m.Regions, region.Id)
	m.regionIds.ReturnId(region.Index)
}

func (m *Model) RangeSortedRegions(f func(id string, region *Region)) {
	var keys []string
	for k := range m.Regions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k, m.Regions[k])
	}
}

type Regions map[string]*Region

type Region struct {
	Scope
	Model       *Model
	Id          string
	Region      string
	Site        string
	Hosts       Hosts
	Index       uint32
	ScaleIndex  uint32
	initialized atomic.Bool
}

func (region *Region) CloneRegion(scaleIndex uint32) *Region {
	result := &Region{
		Scope:      *region.CloneScope(),
		Model:      region.Model,
		Region:     region.Region,
		Site:       region.Site,
		Hosts:      Hosts{},
		Index:      region.Model.GetNextRegionIndex(),
		ScaleIndex: scaleIndex,
	}
	for key, host := range region.Hosts {
		result.Hosts[key] = host.CloneHost(0)
	}
	return result
}

func (region *Region) init(id string, model *Model) {
	if region.initialized.CompareAndSwap(false, true) {
		region.Id = id
		region.Model = model
		region.Index = model.GetNextRegionIndex()
		region.initialize(region, true)
		if region.Data == nil {
			region.Data = Data{}
		}

		if region.Hosts == nil {
			region.Hosts = map[string]*Host{}
		}
	}

	region.RangeSortedHosts(func(id string, host *Host) {
		host.init(id, region)
	})
}

func (region *Region) GetId() string {
	return region.Id
}

func (region *Region) GetType() string {
	return EntityTypeRegion
}

func (region *Region) GetScope() *Scope {
	return &region.Scope
}

func (region *Region) GetModel() *Model {
	return region.Model
}

func (region *Region) GetParentEntity() Entity {
	return region.Model
}

func (region *Region) GetChildren() []Entity {
	if len(region.Hosts) == 0 {
		return nil
	}

	result := make([]Entity, 0, len(region.Hosts))
	for _, entity := range region.Hosts {
		result = append(result, entity)
	}
	return result
}

func (region *Region) Matches(entityType string, matcher EntityMatcher) bool {
	if EntityTypeModel == entityType {
		return region.Model.Matches(entityType, matcher)
	}
	if EntityTypeRegion == entityType {
		return matcher(region)
	}

	if EntityTypeHost == entityType || EntityTypeComponent == entityType {
		for _, child := range region.GetChildren() {
			if child.Matches(entityType, matcher) {
				return true
			}
		}
	}

	return matchHierarchical(entityType, matcher, region)
}

func (region *Region) SelectHosts(hostSpec string) map[string]*Host {
	hosts := map[string]*Host{}
	for id, host := range region.Hosts {
		if hostSpec == "*" || hostSpec == id {
			hosts[id] = host
		} else if strings.HasPrefix(hostSpec, "@") {
			for _, tag := range host.Tags {
				if tag == hostSpec[1:] {
					hosts[id] = host
				}
			}
		}
	}
	return hosts
}

func (region *Region) Accept(visitor EntityVisitor) {
	visitor(region)
	for _, host := range region.Hosts {
		host.Accept(visitor)
	}
}

func (region *Region) RemoveHost(host *Host) {
	delete(region.Hosts, host.Id)
	region.Model.hostIds.ReturnId(host.Index)
}

func (region *Region) RangeSortedHosts(f func(id string, host *Host)) {
	var keys []string
	for k := range region.Hosts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k, region.Hosts[k])
	}
}

type EC2Volume struct {
	Type   string
	SizeGB uint32
	IOPS   uint32
}

type EC2Host struct {
	Volume EC2Volume
}

type Host struct {
	Scope
	Id                   string
	EC2                  EC2Host
	Region               *Region
	PublicIp             string
	PrivateIp            string
	InstanceType         string
	InstanceResourceType string
	SpotPrice            string
	SpotType             string
	Components           Components
	Index                uint32
	ScaleIndex           uint32
	initialized          atomic.Bool
	lock                 sync.Mutex
	sshLock              sync.Mutex
	sshClient            *ssh.Client
	sshConfigFactory     libssh.SshConfigFactory
}

func (host *Host) DoExclusive(f func()) {
	host.lock.Lock()
	defer host.lock.Unlock()
	f()
}

func (host *Host) DoExclusiveFallible(f func() error) error {
	host.lock.Lock()
	defer host.lock.Unlock()
	return f()
}

func (host *Host) ExecLoggedWithTimeout(timeout time.Duration, cmds ...string) (string, error) {
	resultCh := make(chan struct {
		output string
		err    error
	}, 1)

	go func() {
		result, err := host.ExecLogged(cmds...)
		resultCh <- struct {
			output string
			err    error
		}{
			output: result,
			err:    err,
		}
	}()

	select {
	case result := <-resultCh:
		return result.output, result.err
	case <-time.After(timeout):
		return "", errors.Errorf("timed out after %v", timeout)
	}
}

func (host *Host) ExecLogged(cmds ...string) (string, error) {
	buf := &libssh.SyncBuffer{}
	err := host.Exec(buf, cmds...)
	return buf.String(), err
}

func (host *Host) ExecLogOnlyOnError(cmds ...string) error {
	if o, err := host.ExecLogged(cmds...); err != nil {
		logrus.WithField("hostId", host.Id).Errorf("output [%s]", o)
		return fmt.Errorf("error executing process on [%s] (%s)", host.PublicIp, err)
	}
	return nil
}

func (host *Host) GetSshUser() string {
	return host.MustStringVariable("credentials.ssh.username")
}

func (host *Host) NewSshConfigFactory() *libssh.SshConfigFactoryImpl {
	keyPath := host.MustStringVariable("credentials.ssh.key_path")
	return libssh.NewSshConfigFactory(host.GetSshUser(), keyPath, host.PublicIp)
}

func (host *Host) Exec(out io.Writer, cmds ...string) error {
	host.sshLock.Lock()
	defer host.sshLock.Unlock()

	if host.sshClient == nil {
		if host.sshConfigFactory == nil {
			host.sshConfigFactory = host.NewSshConfigFactory()
		}

		client, err := ssh.Dial("tcp", host.sshConfigFactory.Address(), host.sshConfigFactory.Config())
		if err != nil {
			return err
		}
		host.sshClient = client
	}

	for idx, cmd := range cmds {
		session, err := host.sshClient.NewSession()
		if err != nil {
			return err
		}
		session.Stdout = out
		session.Stderr = out

		if idx > 0 {
			logrus.Infof("executing [%s]: '%s'", host.sshConfigFactory.Address(), cmd)
		}
		err = session.Run(cmd)
		_ = session.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func (host *Host) SendFile(localPath string, remotePath string) error {
	localFile, err := os.ReadFile(localPath)

	if err != nil {
		return errors.Wrapf(err, "unable to read local file %v", localFile)
	}

	return host.SendData(localFile, remotePath)
}

func (host *Host) SendData(data []byte, remotePath string) error {
	host.sshLock.Lock()
	defer host.sshLock.Unlock()

	if host.sshClient == nil {
		if host.sshConfigFactory == nil {
			host.sshConfigFactory = host.NewSshConfigFactory()
		}

		client, err := ssh.Dial("tcp", host.sshConfigFactory.Address(), host.sshConfigFactory.Config())
		if err != nil {
			return err
		}
		host.sshClient = client
	}

	client, err := sftp.NewClient(host.sshClient)
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

func (host *Host) FindProcesses(filter func(string) bool) ([]int, error) {
	output, err := host.ExecLogged("ps ax")
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get remote process listing [%s]", host.PublicIp)
	}
	return libssh.FilterProcessList(output, filter)
}

func (host *Host) KillProcesses(signal string, filter func(string) bool) error {
	pidList, err := host.FindProcesses(filter)
	if err != nil {
		return err
	}

	if len(pidList) > 0 {
		killCmd := "sudo kill " + signal + " "
		for _, pid := range pidList {
			killCmd += fmt.Sprintf(" %d", pid)
		}
		killCmd += " || /bin/true"

		output, err := host.ExecLogged(killCmd)
		if err != nil {
			return fmt.Errorf("unable to execute [%v] on [%s] (%s). Output: [%v]", killCmd, host.PublicIp, err, output)
		}
	}

	return nil
}

func (host *Host) CloneHost(scaleIndex uint32) *Host {
	result := &Host{
		Scope:                *host.CloneScope(),
		Id:                   host.Id,
		Region:               host.Region,
		PublicIp:             host.PublicIp,
		PrivateIp:            host.PrivateIp,
		InstanceType:         host.InstanceType,
		InstanceResourceType: host.InstanceResourceType,
		SpotPrice:            host.SpotPrice,
		SpotType:             host.SpotType,
		Components:           Components{},
		Index:                host.Region.Model.GetNextHostIndex(),
		ScaleIndex:           scaleIndex,
	}

	for key, component := range host.Components {
		result.Components[key] = component.CloneComponent(0)
	}

	return result
}

func (host *Host) init(id string, region *Region) {
	logrus.Debugf("initialing host: %v.%v", region.GetId(), id)
	if host.initialized.CompareAndSwap(false, true) {
		host.Id = id
		host.Region = region
		if host.Index == 0 {
			host.Index = region.Model.GetNextHostIndex()
		}
		host.initialize(host, true)
		if host.Data == nil {
			host.Data = Data{}
		}
		if host.Components == nil {
			host.Components = map[string]*Component{}
		}
	}

	host.RangeSortedComponents(func(id string, component *Component) {
		component.init(id, host)
	})
}

func (host *Host) GetId() string {
	return host.Id
}

func (host *Host) GetPath() string {
	return fmt.Sprintf("%v > %v", host.Region.Id, host.Id)
}

func (host *Host) GetType() string {
	return EntityTypeHost
}

func (host *Host) GetScope() *Scope {
	return &host.Scope
}

func (host *Host) GetRegion() *Region {
	return host.Region
}

func (host *Host) GetModel() *Model {
	return host.Region.GetModel()
}

func (host *Host) GetParentEntity() Entity {
	return host.Region
}

func (host *Host) Accept(visitor EntityVisitor) {
	visitor(host)
	for _, component := range host.Components {
		component.Accept(visitor)
	}
}

func (host *Host) GetChildren() []Entity {
	if len(host.Components) == 0 {
		return nil
	}

	result := make([]Entity, 0, len(host.Components))
	for _, entity := range host.Components {
		result = append(result, entity)
	}
	return result
}

func (host *Host) Matches(entityType string, matcher EntityMatcher) bool {
	if EntityTypeModel == entityType || EntityTypeRegion == entityType {
		return host.Region.Matches(entityType, matcher)
	}

	if EntityTypeHost == entityType {
		return matcher(host)
	}

	if EntityTypeComponent == entityType {
		for _, child := range host.GetChildren() {
			if child.Matches(entityType, matcher) {
				return true
			}
		}
	}

	return matchHierarchical(entityType, matcher, host)
}

func (host *Host) RemoveComponent(component *Component) {
	delete(host.Components, component.Id)
	host.GetModel().componentIds.ReturnId(component.Index)
}

func (host *Host) RangeSortedComponents(f func(id string, component *Component)) {
	var keys []string
	for k := range host.Components {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k, host.Components[k])
	}
}

type Hosts map[string]*Host

type Components map[string]*Component

type ActionBinder func(m *Model) Action
type ActionBinders map[string]ActionBinder

func Bind(action Action) ActionBinder {
	return func(m *Model) Action {
		return action
	}
}

func BindF(f func(run Run) error) ActionBinder {
	return Bind(ActionFunc(f))
}

type Action interface {
	Execute(run Run) error
}

type ActionFunc func(run Run) error

func (f ActionFunc) Execute(run Run) error {
	return f(run)
}

func NewRun() (Run, error) {
	result := &runImpl{
		label:          GetLabel(),
		model:          GetModel(),
		runId:          fmt.Sprintf("%d", info.NowInMilliseconds()),
		instanceConfig: instanceConfig,
		oneTimeOps:     cmap.New[*oneTimeOpContext](),
	}
	return result.init()
}

type StagingArea interface {
	GetWorkingDir() string
	GetConfigDir() string
	GetPkiDir() string
	GetBinDir() string
	GetTmpDir() string

	DirExists(path string) (bool, error)
	FileExists(path string) (bool, error)

	DoOnce(operation string, f func() error) error
}

type Run interface {
	StagingArea
	GetModel() *Model
	GetLabel() *Label
	GetId() string
}

type runImpl struct {
	label          *Label
	model          *Model
	runId          string
	instanceConfig *InstanceConfig
	oneTimeOps     cmap.ConcurrentMap[string, *oneTimeOpContext]
}

func (self *runImpl) DoOnce(operation string, f func() error) error {
	var ctx *oneTimeOpContext
	ctx, found := self.oneTimeOps.Get(operation)
	opOwner := false
	if !found {
		ctx = newOneTimeOpContext()
		if opOwner = self.oneTimeOps.SetIfAbsent(operation, ctx); !opOwner {
			ctx, _ = self.oneTimeOps.Get(operation)
		}
	}
	if opOwner {
		return ctx.runOp(f)
	}
	return ctx.getOpResult()
}

func (self *runImpl) init() (*runImpl, error) {
	if err := os.MkdirAll(self.GetBinDir(), 0700); err != nil {
		return nil, errors.Wrapf(err, "unable to create binaries working directory [%s]", self.GetBinDir())
	}
	if err := os.MkdirAll(self.GetConfigDir(), 0700); err != nil {
		return nil, errors.Wrapf(err, "unable to create config working directory [%s]", self.GetConfigDir())
	}
	if err := os.MkdirAll(self.GetPkiDir(), 0700); err != nil {
		return nil, errors.Wrapf(err, "unable to create pki working directory [%s]", self.GetPkiDir())
	}
	if err := os.MkdirAll(self.GetTmpDir(), 0700); err != nil {
		return nil, errors.Wrapf(err, "unable to create tmp working directory [%s]", self.GetTmpDir())
	}

	// ensure a cfg dir exists for things like re-enrollment JWTs
	if err := os.MkdirAll(filepath.Join(self.GetWorkingDir(), BuildConfigDir), 0700); err != nil {
		return nil, errors.Wrapf(err, "unable to create config working directory [%s]", self.GetConfigDir())
	}

	return self, nil
}

func (self *runImpl) GetWorkingDir() string {
	return self.instanceConfig.WorkingDirectory
}

func (self *runImpl) GetConfigDir() string {
	return filepath.Join(self.GetWorkingDir(), BuildKitDir, BuildConfigDir)
}

func (self *runImpl) GetPkiDir() string {
	return filepath.Join(self.GetWorkingDir(), BuildKitDir, BuildPkiDir)
}

func (self *runImpl) GetBinDir() string {
	return filepath.Join(self.GetWorkingDir(), BuildKitDir, BuildBinDir)
}

func (self *runImpl) GetTmpDir() string {
	return filepath.Join(self.GetWorkingDir(), BuildTmpDir)
}

func (self *runImpl) DirExists(path string) (bool, error) {
	fullPath := filepath.Join(self.GetWorkingDir(), path)
	s, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, nil
	}
	return s.IsDir(), nil
}

func (self *runImpl) FileExists(path string) (bool, error) {
	fullPath := filepath.Join(self.instanceConfig.WorkingDirectory, path)
	s, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, nil
	}
	return !s.IsDir(), nil
}

func (self *runImpl) GetModel() *Model {
	return self.model
}

func (self *runImpl) GetLabel() *Label {
	return self.label
}

func (self *runImpl) GetId() string {
	return self.runId
}

func newOneTimeOpContext() *oneTimeOpContext {
	return &oneTimeOpContext{
		doneC: make(chan struct{}),
	}
}

type oneTimeOpContext struct {
	doneC chan struct{}
	err   error
	sync.Mutex
}

func (self *oneTimeOpContext) runOp(f func() error) error {
	result := f()
	self.Lock()
	defer self.Unlock()
	self.err = result
	close(self.doneC)
	return self.err
}

func (self *oneTimeOpContext) getOpResult() error {
	<-self.doneC
	return self.Err()
}

func (self *oneTimeOpContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (self *oneTimeOpContext) Done() <-chan struct{} {
	return self.doneC
}

func (self *oneTimeOpContext) Err() error {
	self.Lock()
	defer self.Unlock()
	return self.err
}

func (self *oneTimeOpContext) Value(any) any {
	return nil
}

type Stages []Stage

type Stage interface {
	Execute(run Run) error
}

type StageActionF func(run Run) error

func (self StageActionF) Execute(run Run) error {
	return self(run)
}

func RunAction(action string) Action {
	return actionStage(action)
}

type actionStage string

func (stage actionStage) Execute(run Run) error {
	return stage.execute(run)
}

func (stage actionStage) execute(run Run) error {
	actionName := string(stage)
	m := run.GetModel()
	action, found := m.GetAction(actionName)
	if !found {
		return fmt.Errorf("no [%s] action", actionName)
	}
	figlet.FigletMini("action: " + actionName)
	if err := action.Execute(run); err != nil {
		return fmt.Errorf("error executing [%s] action (%w)", actionName, err)
	}
	return nil
}

func (m *Model) AddActionBinder(actionName string, action ActionBinder) {
	m.Actions[actionName] = action
}

func (m *Model) AddAction(actionName string, action Action) {
	m.Actions[actionName] = Bind(action)
}

func (m *Model) AddActionF(actionName string, action ActionFunc) {
	m.Actions[actionName] = Bind(action)
}

func (m *Model) ExecuteAction(actionName string) Action {
	return ActionFunc(func(run Run) error {
		action, found := m.GetAction(actionName)
		if !found {
			return fmt.Errorf("no '%s' action defined", actionName)
		}
		figlet.FigletMini("action: " + actionName)
		return action.Execute(run)
	})
}

func (m *Model) AddActivationStage(stage Stage) {
	m.Activation = append(m.Activation, stage)
}

func (m *Model) AddActivationStageF(stage StageActionF) {
	m.Activation = append(m.Activation, stage)
}

func (m *Model) AddActivationStages(stage ...Stage) {
	m.Activation = append(m.Activation, stage...)
}

func (m *Model) AddActivationActions(actions ...string) {
	for _, action := range actions {
		m.AddActivationStage(actionStage(action))
	}
}

func (m *Model) AddOperatingStage(stage Stage) {
	m.Operation = append(m.Operation, stage)
}

func (m *Model) AddOperatingStageF(stage StageActionF) {
	m.AddOperatingStage(stage)
}

func (m *Model) AddOperatingStages(stages ...Stage) {
	m.Operation = append(m.Operation, stages...)
}

func (m *Model) AddOperatingActions(actions ...string) {
	for _, action := range actions {
		m.AddOperatingStage(actionStage(action))
	}
}

func (m *Model) Express(run Run) error {
	for _, stage := range m.Infrastructure {
		if err := stage.Execute(run); err != nil {
			return fmt.Errorf("error expressing infrastructure (%w)", err)
		}
	}
	run.GetLabel().State = Expressed
	if err := run.GetLabel().Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) Build(run Run) error {
	err := m.ForEachComponent("*", 1, func(c *Component) error {
		if stageable, ok := c.Type.(FileStagingComponent); ok {
			return stageable.StageFiles(run, c)
		}
		return nil
	})

	if err != nil {
		return err
	}

	for _, stage := range m.Configuration {
		if err := stage.Execute(run); err != nil {
			return fmt.Errorf("error building configuration (%w)", err)
		}
	}
	run.GetLabel().State = Configured
	if err := run.GetLabel().Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) Sync(run Run) error {
	for idx, stage := range m.Distribution {
		if err := stage.Execute(run); err != nil {
			return fmt.Errorf("error distributing stage %d - %T, (%w)", idx+1, stage, err)
		}
	}

	err := m.ForEachHost("*", 100, func(host *Host) error {
		for _, c := range host.Components {
			hostInitializer, ok := c.Type.(HostInitializingComponent)
			if !ok {
				continue
			}

			if err := hostInitializer.InitializeHost(run, c); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	run.GetLabel().State = Distributed
	if err := run.GetLabel().Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}

	return nil
}

func (m *Model) Activate(run Run) error {
	for _, stage := range m.Activation {
		if err := stage.Execute(run); err != nil {
			return fmt.Errorf("error activating (%w)", err)
		}
	}
	run.GetLabel().State = Activated
	if err := run.GetLabel().Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) Operate(run Run) error {
	for _, stage := range m.Operation {
		if err := stage.Execute(run); err != nil {
			return fmt.Errorf("error operating (%w)", err)
		}
	}
	run.GetLabel().State = Operating
	if err := run.GetLabel().Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) Dispose(run Run) error {
	for _, stage := range m.Disposal {
		if err := stage.Execute(run); err != nil {
			return fmt.Errorf("error disposing (%w)", err)
		}
	}
	run.GetLabel().State = Disposed
	if err := run.GetLabel().Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) AcceptHostMetrics(host *Host, event *MetricsEvent) {
	for _, handler := range m.MetricsHandlers {
		handler.AcceptHostMetrics(host, event)
	}
}

func GetScopedEntityPath(entity Entity) []string {
	parent := entity.GetParentEntity()
	if parent != nil {
		// dont' want to include the model in the path
		if _, isModel := parent.(*Model); !isModel {
			return append(GetScopedEntityPath(parent), entity.GetId())
		}
	}
	return []string{entity.GetId()}
}
