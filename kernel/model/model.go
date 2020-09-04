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
	Regions Regions

	Factories           []Factory
	BootstrapExtensions []BootstrapExtension
	Actions             map[string]ActionBinder
	Infrastructure      InfrastructureBinders
	Configuration       ConfigurationBinders
	Kitting             KittingBinders
	Distribution        DistributionBinders
	Activation          ActivationBinders
	Operation           OperatingBinders
	Disposal            DisposalBinders

	actions              map[string]Action
	infrastructureStages []InfrastructureStage
	configurationStages  []ConfigurationStage
	kittingStages        []KittingStage
	distributionStages   []DistributionStage
	activationStages     []ActivationStage
	operationStages      []OperatingStage
	disposalStages       []DisposalStage

	initialized concurrenz.AtomicBoolean
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
		for id, region := range m.Regions {
			region.init(id, m)
		}

		// trim tag prefixes
		m.Accept(func(e Entity) {
			var tags Tags
			for _, tag := range e.GetScope().Tags {
				tag = strings.TrimPrefix(tag, DontInheritTagPrefix)
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
	model  *Model
	id     string
	Region string
	Site   string
	Hosts  Hosts
}

func (region *Region) init(id string, model *Model) {
	region.id = id
	region.model = model
	region.Scope.setParent(&model.Scope)

	for hostId, host := range region.Hosts {
		host.init(hostId, region)
	}
}

func (region *Region) GetId() string {
	return region.id
}

func (region *Region) GetType() string {
	return EntityTypeRegion
}

func (region *Region) GetScope() *Scope {
	return &region.Scope
}

func (region *Region) GetModel() *Model {
	return region.model
}

func (region *Region) GetParentEntity() Entity {
	return region.model
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
		return region.model.Matches(entityType, matcher)
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
	id                   string
	region               *Region
	PublicIp             string
	PrivateIp            string
	InstanceType         string
	InstanceResourceType string
	SpotPrice            string
	SpotType             string
	Components           Components
}

func (host *Host) init(id string, region *Region) {
	logrus.Infof("initialing host: %v.%v", region.GetId(), id)
	host.id = id
	host.region = region
	host.Scope.setParent(&region.Scope)

	for componentId, component := range host.Components {
		component.init(componentId, host)
	}
}

func (host *Host) GetId() string {
	return host.id
}

func (host *Host) GetType() string {
	return EntityTypeHost
}

func (host *Host) GetScope() *Scope {
	return &host.Scope
}

func (host *Host) GetRegion() *Region {
	return host.region
}

func (host *Host) GetParentEntity() Entity {
	return host.region
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
		return host.region.Matches(entityType, matcher)
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
	id              string
	host            *Host
	ScriptSrc       string
	ScriptName      string
	ConfigSrc       string
	ConfigName      string
	BinaryName      string
	PublicIdentity  string
	PrivateIdentity string
}

func (component *Component) init(id string, host *Host) {
	component.id = id
	component.Scope.setParent(&host.Scope)
	component.host = host
}

func (component *Component) GetId() string {
	return component.id
}

func (component *Component) GetType() string {
	return EntityTypeComponent
}

func (component *Component) GetScope() *Scope {
	return &component.Scope
}

func (component *Component) GetHost() *Host {
	return component.host
}

func (component *Component) GetRegion() *Region {
	return component.host.region
}

func (component *Component) GetModel() *Model {
	return component.host.region.model
}

func (component *Component) GetParentEntity() Entity {
	return component.host
}

func (component *Component) Accept(visitor EntityVisitor) {
	visitor(component)
}

func (component *Component) GetChildren() []Entity {
	return nil
}

func (component *Component) Matches(entityType string, matcher EntityMatcher) bool {
	if EntityTypeModel == entityType || EntityTypeRegion == entityType || EntityTypeHost == entityType {
		return component.host.Matches(entityType, matcher)
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

type InfrastructureStage interface {
	Express(m *Model, l *Label) error
}

type ConfigurationStage interface {
	Configure(m *Model) error
}

type KittingStage interface {
	Kit(m *Model) error
}

type DistributionStage interface {
	Distribute(m *Model) error
}

type ActivationStage interface {
	Activate(m *Model) error
}

type OperatingStage interface {
	Operate(m *Model, run string) error
}

type DisposalStage interface {
	Dispose(m *Model) error
}

type InfrastructureBinder func(m *Model) InfrastructureStage
type InfrastructureBinders []InfrastructureBinder

type ConfigurationBinder func(m *Model) ConfigurationStage
type ConfigurationBinders []ConfigurationBinder

type KittingBinder func(m *Model) KittingStage
type KittingBinders []KittingBinder

type DistributionBinder func(m *Model) DistributionStage
type DistributionBinders []DistributionBinder

type ActivationBinder func(m *Model) ActivationStage
type ActivationBinders []ActivationBinder

type OperatingBinder func(m *Model) OperatingStage
type OperatingBinders []OperatingBinder

type DisposalBinder func(m *Model) DisposalStage
type DisposalBinders []DisposalBinder

func (m *Model) Express(l *Label) error {
	for _, stage := range m.infrastructureStages {
		if err := stage.Express(m, l); err != nil {
			return fmt.Errorf("error expressing infrastructure (%w)", err)
		}
	}
	l.State = Expressed
	if err := l.Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) Build(l *Label) error {
	for _, stage := range m.configurationStages {
		if err := stage.Configure(m); err != nil {
			return fmt.Errorf("error building configuration (%w)", err)
		}
	}
	l.State = Configured
	if err := l.Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) Kit(l *Label) error {
	for _, stage := range m.kittingStages {
		if err := stage.Kit(m); err != nil {
			return fmt.Errorf("error kitting (%w)", err)
		}
	}
	l.State = Kitted
	if err := l.Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) Sync(l *Label) error {
	for _, stage := range m.distributionStages {
		if err := stage.Distribute(m); err != nil {
			return fmt.Errorf("error distributing (%w)", err)
		}
	}
	l.State = Distributed
	if err := l.Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) Activate(l *Label) error {
	for _, stage := range m.activationStages {
		if err := stage.Activate(m); err != nil {
			return fmt.Errorf("error activating (%w)", err)
		}
	}
	l.State = Activated
	if err := l.Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) Operate(l *Label) error {
	run := fmt.Sprintf("%d", info.NowInMilliseconds())
	for _, stage := range m.operationStages {
		if err := stage.Operate(m, run); err != nil {
			return fmt.Errorf("error operating (%w)", err)
		}
	}
	l.State = Operating
	if err := l.Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}

func (m *Model) Dispose(l *Label) error {
	for _, stage := range m.disposalStages {
		if err := stage.Dispose(m); err != nil {
			return fmt.Errorf("error disposing (%w)", err)
		}
	}
	l.State = Disposed
	if err := l.Save(); err != nil {
		return fmt.Errorf("error updating instance label (%w)", err)
	}
	return nil
}
