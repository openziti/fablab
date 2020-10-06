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

package model

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib/figlet"
	"github.com/openziti/foundation/util/concurrenz"
	"github.com/openziti/foundation/util/info"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	EntityTypeModel     = "model"
	EntityTypeRegion    = "region"
	EntityTypeHost      = "host"
	EntityTypeComponent = "component"
)

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
}

type Model struct {
	name   string
	Parent *Model

	Scope
	Regions             Regions
	ModelFactories      []Factory // Factories that change the model structure, eg: add/remove hosts
	Factories           []Factory
	BootstrapExtensions []BootstrapExtension
	Actions             map[string]ActionBinder
	Infrastructure      InfrastructureStages
	Configuration       ConfigurationStages
	Distribution        DistributionStages
	Activation          ActivationStages
	Operation           OperatingStages
	Disposal            DisposalStages
	MetricsHandlers     []MetricsHandler

	actions map[string]Action

	initialized concurrenz.AtomicBoolean
}

func (m *Model) GetModel() *Model {
	return m
}

func (m *Model) GetId() string {
	return m.name
}

func (m *Model) GetType() string {
	return EntityTypeModel
}

func (m *Model) GetScope() *Scope {
	return &m.Scope
}

func (m *Model) GetParentEntity() Entity {
	return m.Parent
}

func (m *Model) Matches(entityType string, matcher EntityMatcher) bool {
	if EntityTypeModel == entityType {
		return matcher(m) || (m.Parent != nil && m.Parent.Matches(entityType, matcher))
	}

	if EntityTypeRegion == entityType || EntityTypeHost == entityType || EntityTypeComponent == entityType {
		for _, child := range m.GetChildren() {
			if child.Matches(entityType, matcher) {
				return true
			}
		}
	}

	return false
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

func (m *Model) init(name string) {
	if m.initialized.CompareAndSwap(false, true) {
		m.name = name
		if m.Data == nil {
			m.Data = Data{}
		}
		for id, region := range m.Regions {
			region.init(id, m)
		}

		// trim tag prefixes
		m.Accept(func(e Entity) {
			var tags Tags
			for _, tag := range e.GetScope().Tags {
				tag = strings.TrimPrefix(tag, InheritTagPrefix)
				tags = append(tags, tag)
			}
			e.GetScope().Tags = tags
		})
	}
}

func (m *Model) Accept(visitor EntityVisitor) {
	visitor(m)
	for _, region := range m.Regions {
		region.Accept(visitor)
	}
}

type Regions map[string]*Region

type Region struct {
	Scope
	Model  *Model
	Id     string
	Region string
	Site   string
	Hosts  Hosts
	Index  int
}

func (region *Region) CloneRegion(index int) *Region {
	result := &Region{
		Scope:  *region.CloneScope(),
		Model:  region.Model,
		Region: region.Region,
		Site:   region.Site,
		Hosts:  Hosts{},
		Index:  index,
	}
	for key, host := range region.Hosts {
		result.Hosts[key] = host.CloneHost(0)
	}
	return result
}

func (region *Region) Templatize(templater *Templater) {
	region.Scope.Templatize(templater)
	region.Region = templater.Templatize(region.Region)
	region.Site = templater.Templatize(region.Site)
}

func (region *Region) init(id string, model *Model) {
	region.Id = id
	region.Model = model
	region.Scope.setParent(&model.Scope)
	if region.Data == nil {
		region.Data = Data{}
	}
	for hostId, host := range region.Hosts {
		host.init(hostId, region)
	}
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
	return false
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

type Host struct {
	Scope
	Id                   string
	Region               *Region
	PublicIp             string
	PrivateIp            string
	InstanceType         string
	InstanceResourceType string
	SpotPrice            string
	SpotType             string
	Components           Components
	Index                int
}

func (host *Host) CloneHost(index int) *Host {
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
		Index:                index,
	}

	for key, component := range host.Components {
		result.Components[key] = component.CloneComponent(0)
	}

	return result
}

func (host *Host) Templatize(templater *Templater) {
	host.Scope.Templatize(templater)
	host.PublicIp = templater.Templatize(host.PublicIp)
	host.PrivateIp = templater.Templatize(host.PrivateIp)
	host.InstanceType = templater.Templatize(host.InstanceType)
	host.InstanceResourceType = templater.Templatize(host.InstanceResourceType)
	host.SpotPrice = templater.Templatize(host.SpotPrice)
	host.SpotType = templater.Templatize(host.SpotType)
}

func (host *Host) init(id string, region *Region) {
	logrus.Debugf("initialing host: %v.%v", region.GetId(), id)
	host.Id = id
	host.Region = region
	host.Scope.setParent(&region.Scope)
	if host.Data == nil {
		host.Data = Data{}
	}
	for componentId, component := range host.Components {
		component.init(componentId, host)
	}
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
	return false
}

type Hosts map[string]*Host

type Component struct {
	Scope
	Id              string
	Host            *Host
	ScriptSrc       string
	ScriptName      string
	ConfigSrc       string
	ConfigName      string
	BinaryName      string
	PublicIdentity  string
	PrivateIdentity string
	Index           int
}

func (component *Component) CloneComponent(index int) *Component {
	result := &Component{
		Scope:           *component.CloneScope(),
		Id:              component.Id,
		Host:            component.Host,
		ScriptSrc:       component.ScriptSrc,
		ScriptName:      component.ScriptName,
		ConfigSrc:       component.ConfigSrc,
		ConfigName:      component.ConfigName,
		BinaryName:      component.BinaryName,
		PublicIdentity:  component.PublicIdentity,
		PrivateIdentity: component.PrivateIdentity,
		Index:           index,
	}
	return result
}

func (component *Component) Templatize(templater *Templater) {
	component.Scope.Templatize(templater)
	component.ScriptSrc = templater.Templatize(component.ScriptSrc)
	component.ScriptName = templater.Templatize(component.ScriptName)
	component.ConfigSrc = templater.Templatize(component.ConfigSrc)
	component.ConfigName = templater.Templatize(component.ConfigName)
	component.BinaryName = templater.Templatize(component.BinaryName)
	component.PublicIdentity = templater.Templatize(component.PublicIdentity)
	component.PrivateIdentity = templater.Templatize(component.PrivateIdentity)
}

func (component *Component) init(id string, host *Host) {
	component.Id = id
	component.Scope.setParent(&host.Scope)
	component.Host = host
	if component.Data == nil {
		component.Data = Data{}
	}
}

func (component *Component) GetId() string {
	return component.Id
}

func (component *Component) GetPath() string {
	return fmt.Sprintf("%v > %v", component.Host.GetPath(), component.Id)
}

func (component *Component) GetType() string {
	return EntityTypeComponent
}

func (component *Component) GetScope() *Scope {
	return &component.Scope
}

func (component *Component) GetHost() *Host {
	return component.Host
}

// Region is used by templates, so we can reference via {{ .Region.Field }}
func (component *Component) Region() *Region {
	return component.Host.Region
}

func (component *Component) GetRegion() *Region {
	return component.Host.Region
}

func (component *Component) GetModel() *Model {
	return component.Host.Region.Model
}

func (component *Component) GetParentEntity() Entity {
	return component.Host
}

func (component *Component) Accept(visitor EntityVisitor) {
	visitor(component)
}

func (component *Component) GetChildren() []Entity {
	return nil
}

func (component *Component) Matches(entityType string, matcher EntityMatcher) bool {
	if EntityTypeModel == entityType || EntityTypeRegion == entityType || EntityTypeHost == entityType {
		return component.Host.Matches(entityType, matcher)
	}
	if EntityTypeComponent == entityType {
		return matcher(component)
	}
	return false
}

type Components map[string]*Component

type ActionBinder func(m *Model) Action
type ActionBinders map[string]ActionBinder

type Action interface {
	Execute(m *Model) error
}

type ActionFunc func(m *Model) error

func (f ActionFunc) Execute(m *Model) error {
	return f(m)
}

func NewRun(label *Label, model *Model) Run {
	return &runImpl{
		label: label,
		model: model,
		runId: fmt.Sprintf("%d", info.NowInMilliseconds()),
	}
}

type Run interface {
	GetModel() *Model
	GetLabel() *Label
	GetId() string
}

type runImpl struct {
	label *Label
	model *Model
	runId string
}

func (run *runImpl) GetModel() *Model {
	return run.model
}

func (run *runImpl) GetLabel() *Label {
	return run.label
}

func (run *runImpl) GetId() string {
	return run.runId
}

type InfrastructureStages []InfrastructureStage

type InfrastructureStage interface {
	Express(run Run) error
}

type ConfigurationStages []ConfigurationStage

type ConfigurationStage interface {
	Configure(run Run) error
}

type DistributionStages []DistributionStage

type DistributionStage interface {
	Distribute(run Run) error
}

type ActivationStages []ActivationStage

type ActivationStage interface {
	Activate(run Run) error
}

type OperatingStages []OperatingStage

type OperatingStage interface {
	Operate(run Run) error
}

type DisposalStages []DisposalStage

type DisposalStage interface {
	Dispose(run Run) error
}

type actionStage string

func (stage actionStage) Activate(run Run) error {
	return stage.execute(run)
}

func (stage actionStage) Operate(run Run) error {
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
	if err := action.Execute(m); err != nil {
		return fmt.Errorf("error executing [%s] action (%w)", actionName, err)
	}
	return nil
}

func (m *Model) AddActivationStage(stage ActivationStage) {
	m.Activation = append(m.Activation, stage)
}

func (m *Model) AddActivationStages(stage ...ActivationStage) {
	m.Activation = append(m.Activation, stage...)
}

func (m *Model) AddActivationActions(actions ...string) {
	for _, action := range actions {
		m.AddActivationStage(actionStage(action))
	}
}

func (m *Model) AddOperatingStage(stage OperatingStage) {
	m.Operation = append(m.Operation, stage)
}

func (m *Model) AddOperatingStages(stages ...OperatingStage) {
	m.Operation = append(m.Operation, stages...)
}

func (m *Model) AddOperatingActions(actions ...string) {
	for _, action := range actions {
		m.AddOperatingStage(actionStage(action))
	}
}

func (m *Model) Express(run Run) error {
	for _, stage := range m.Infrastructure {
		if err := stage.Express(run); err != nil {
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
	for _, stage := range m.Configuration {
		if err := stage.Configure(run); err != nil {
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
	for _, stage := range m.Distribution {
		if err := stage.Distribute(run); err != nil {
			return fmt.Errorf("error distributing (%w)", err)
		}
	}
	run.GetLabel().State = Distributed
	if err := run.GetLabel().Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) Activate(run Run) error {
	for _, stage := range m.Activation {
		if err := stage.Activate(run); err != nil {
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
		if err := stage.Operate(run); err != nil {
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
		if err := stage.Dispose(run); err != nil {
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
