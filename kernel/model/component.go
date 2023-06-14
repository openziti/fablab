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

package model

import (
	"fmt"
	"sync/atomic"
)

const (
	ComponentActionStop           = "stop"
	ComponentActionStart          = "start"
	ComponentActionStageFiles     = "stageFiles"
	ComponentActionInitializeHost = "initializeHost"
)

// ComponentType contains the custom logic for a component. This can
// range from provisioning to configuration to running
type ComponentType interface {
	// Dump returns a JSON marshallable object allowing the strategy data to be dumped for inspection
	Dump() any

	// IsRunning returns true if the component is currently represented by a running process, false otherwise
	IsRunning(run Run, c *Component) (bool, error)

	// Stop will stop any currently running processes which represent the component
	Stop(run Run, c *Component) error
}

// A ServerComponent is one which can be started and left running in the background
type ServerComponent interface {
	Start(run Run, c *Component) error
}

// A FileStagingComponent is able to contribute files to the staging area, to be synced
// up to the components host. This may include things like binaries, scripts, configuration
// files and PKI.
type FileStagingComponent interface {
	ComponentType

	// StageFiles is called at the beginning of the configuration phase and allows the component to contribute
	// files to be synced to the Host
	StageFiles(r Run, c *Component) error
}

// A HostInitializingComponent can run some one-time configuration on the host as part of
// the distibution/sync phase. This would include things like adjusting system configuration
// files on the host.
type HostInitializingComponent interface {
	ComponentType

	// InitializeHost is called at the end of the distribution phase and allows the component to
	// make changes to Host configuration
	InitializeHost(r Run, c *Component) error
}

// A ComponentAction is an action execute in the context of a specific component
type ComponentAction interface {
	Execute(r Run, c *Component) error
}

// ComponentActionF is the function version of ComponentAction
type ComponentActionF func(r Run, c *Component) error

func (f ComponentActionF) Execute(r Run, c *Component) error {
	return f(r, c)
}

// An ActionsComponent provides additional actions which can be executed using the ExecuteAction method
type ActionsComponent interface {
	ComponentType

	// GetActions returns the set of additional actions available on the component
	GetActions() map[string]ComponentAction
}

type Component struct {
	Scope
	Id          string
	Host        *Host
	Type        ComponentType
	Index       uint32
	ScaleIndex  uint32
	initialized atomic.Bool
}

func (component *Component) CloneComponent(scaleIndex uint32) *Component {
	result := &Component{
		Scope:      *component.CloneScope(),
		Id:         component.Id,
		Type:       component.Type,
		Host:       component.Host,
		Index:      component.GetModel().GetNextComponentIndex(),
		ScaleIndex: scaleIndex,
	}
	return result
}

func (component *Component) init(id string, host *Host) {
	if component.initialized.CompareAndSwap(false, true) {
		component.Id = id
		component.Host = host
		component.Index = host.GetModel().GetNextComponentIndex()
		component.Scope.initialize(component, true)
		if component.Data == nil {
			component.Data = Data{}
		}
	}
}

func (component *Component) GetId() string {
	return component.Id
}

func (component *Component) GetPath() string {
	return fmt.Sprintf("%v > %v", component.Host.GetPath(), component.Id)
}

func (component *Component) GetPathId() string {
	return component.GetRegion().Id + "." + component.Host.Id + "." + component.Id
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

	return matchHierarchical(entityType, matcher, component)
}
