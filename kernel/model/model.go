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
	"strings"
)

type Entity interface {
	GetId() string
	GetScope() *Scope
	GetParentEntity() Entity
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

func (m *Model) GetScope() *Scope {
	return &m.Scope
}

func (m *Model) GetParentEntity() Entity {
	return m.Parent
}

func (m *Model) init(name string) {
	m.name = name
	if m.initialized.CompareAndSwap(false, true) {
		for id, region := range m.Regions {
			region.init(id, m)
		}
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

func (region *Region) GetScope() *Scope {
	return &region.Scope
}

func (region *Region) GetModel() *Model {
	return region.model
}

func (region *Region) GetParentEntity() Entity {
	return region.model
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

func (host *Host) GetScope() *Scope {
	return &host.Scope
}

func (host *Host) GetRegion() *Region {
	return host.region
}

func (host *Host) GetParentEntity() Entity {
	return host.region
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
