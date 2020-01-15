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
)

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
	for _, stage := range m.operationStages {
		if err := stage.Operate(m); err != nil {
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

type Model struct {
	Parent *Model

	Scope
	Regions Regions

	Factories      []Factory
	Actions        map[string]ActionBinder
	Infrastructure InfrastructureBinders
	Configuration  ConfigurationBinders
	Kitting        KittingBinders
	Distribution   DistributionBinders
	Activation     ActivationBinders
	Operation      OperatingBinders
	Disposal       DisposalBinders

	actions              map[string]Action
	infrastructureStages []InfrastructureStage
	configurationStages  []ConfigurationStage
	kittingStages        []KittingStage
	distributionStages   []DistributionStage
	activationStages     []ActivationStage
	operationStages      []OperatingStage
	disposalStages       []DisposalStage
}

type Regions map[string]*Region

type Region struct {
	Scope
	Id    string
	Az    string
	Hosts Hosts
}

type Host struct {
	Scope
	PublicIp     string
	PrivateIp    string
	InstanceType string
	Components   Components
}

type Hosts map[string]*Host

type Component struct {
	Scope
	ScriptSrc       string
	ScriptName      string
	ConfigSrc       string
	ConfigName      string
	BinaryName      string
	PublicIdentity  string
	PrivateIdentity string
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
	Operate(m *Model) error
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
