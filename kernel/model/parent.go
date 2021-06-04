/*
	Copyright 2020 NetFoundry, Inc.

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
	"github.com/jinzhu/copier"
)

func (m *Model) Merge(parent *Model) error {
	var err error
	if m.Scope, err = m.Scope.Merge(parent.Scope); err != nil {
		return fmt.Errorf("error merging model scope (%w)", err)
	}
	if m.Regions, err = m.Regions.Merge(parent.Regions); err != nil {
		return fmt.Errorf("error merging model regions (%w)", err)
	}
	if err = m.mergeActionsAndStages(parent); err != nil {
		return fmt.Errorf("error merging actions and binders (%w)", err)
	}
	return nil
}

func (m *Model) mergeActionsAndStages(parent *Model) error {
	mergedFactories := make([]Factory, 0)
	if err := copier.Copy(&mergedFactories, parent.Factories); err != nil {
		return fmt.Errorf("error copying parent factories (%w)", err)
	}
	mergedFactories = append(mergedFactories, m.Factories...)
	m.Factories = mergedFactories

	mergedActions := make(map[string]ActionBinder)
	if err := copier.Copy(&mergedActions, parent.Actions); err != nil {
		return fmt.Errorf("error copying parent action binders (%w)", err)
	}
	for k, v := range m.Actions {
		mergedActions[k] = v
	}
	m.Actions = mergedActions

	mergedInfrastructure := make([]InfrastructureStage, 0)
	if err := copier.Copy(&mergedInfrastructure, parent.Infrastructure); err != nil {
		return fmt.Errorf("error copying parent infrastructure binders (%w)", err)
	}
	mergedInfrastructure = append(mergedInfrastructure, m.Infrastructure...)
	m.Infrastructure = mergedInfrastructure

	mergedConfiguration := make([]ConfigurationStage, 0)
	if err := copier.Copy(&mergedConfiguration, parent.Configuration); err != nil {
		return fmt.Errorf("error copying parent configuration binders (%w)", err)
	}
	mergedConfiguration = append(mergedConfiguration, m.Configuration...)
	m.Configuration = mergedConfiguration

	mergedDistribution := make([]DistributionStage, 0)
	if err := copier.Copy(&mergedDistribution, parent.Distribution); err != nil {
		return fmt.Errorf("error copying parent distribution binders (%w)", err)
	}
	mergedDistribution = append(mergedDistribution, m.Distribution...)
	m.Distribution = mergedDistribution

	mergedActivation := make([]ActivationStage, 0)
	if err := copier.Copy(&mergedActions, parent.Activation); err != nil {
		return fmt.Errorf("error copying parent activation binders (%w)", err)
	}
	mergedActivation = append(mergedActivation, m.Activation...)
	m.Activation = mergedActivation

	mergedOperation := make([]OperatingStage, 0)
	if err := copier.Copy(&mergedOperation, parent.Operation); err != nil {
		return fmt.Errorf("error copying parent operation binders (%w)", err)
	}
	mergedOperation = append(mergedOperation, m.Operation...)
	m.Operation = mergedOperation

	mergedDisposal := make([]DisposalStage, 0)
	if err := copier.Copy(&mergedDisposal, parent.Disposal); err != nil {
		return fmt.Errorf("error copying parent disposal binders (%w)", err)
	}
	mergedDisposal = append(mergedDisposal, m.Disposal...)
	m.Disposal = mergedDisposal

	return nil
}

func (s Scope) Merge(parent Scope) (Scope, error) {
	merged := Scope{}
	if err := copier.Copy(&merged, parent); err != nil {
		return Scope{}, fmt.Errorf("error copying parent (%w)", err)
	}

	for k, v := range s.Defaults {
		if merged.Defaults == nil {
			merged.Defaults = make(Variables)
		}
		merged.Defaults[k] = v
	}

	for k, v := range s.Data {
		if merged.Data == nil {
			merged.Data = make(Data)
		}
		merged.Data[k] = v
	}

	merged.Tags = append(merged.Tags, s.Tags...)
	merged.entity = s.entity
	return merged, nil
}

func (r Regions) Merge(parent Regions) (Regions, error) {
	merged := make(Regions)
	if err := copier.Copy(&merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	for k, v := range r {
		if pv, found := merged[k]; found {
			if mergedRegion, err := v.Merge(pv); err == nil {
				merged[k] = mergedRegion
			} else {
				return nil, fmt.Errorf("error merging region [%s] (%w)", k, err)
			}
		} else {
			merged[k] = v
		}
	}

	return merged, nil
}

func (r *Region) Merge(parent *Region) (*Region, error) {
	merged := &Region{}
	if err := copier.Copy(merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	var err error
	if merged.Scope, err = r.Scope.Merge(parent.Scope); err != nil {
		return nil, fmt.Errorf("error merging region scope (%w)", err)
	}

	if r.Region != "" {
		merged.Region = r.Region
	}
	if r.Site != "" {
		merged.Site = r.Site
	}

	if merged.Hosts, err = r.Hosts.Merge(parent.Hosts); err != nil {
		return nil, fmt.Errorf("error merging hosts (%w)", err)
	}

	return merged, nil
}

func (h Hosts) Merge(parent Hosts) (Hosts, error) {
	merged := make(Hosts)
	if err := copier.Copy(&merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	for k, v := range h {
		if pv, found := merged[k]; found {
			if mergedHost, err := v.Merge(pv); err == nil {
				merged[k] = mergedHost
			} else {
				return nil, fmt.Errorf("error merging host (%w)", err)
			}
		} else {
			merged[k] = v
		}
	}

	return merged, nil
}

func (h *Host) Merge(parent *Host) (*Host, error) {
	merged := &Host{}
	if err := copier.Copy(merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	var err error
	if merged.Scope, err = h.Scope.Merge(parent.Scope); err != nil {
		return nil, fmt.Errorf("error merging host scope (%w)", err)
	}

	if h.PublicIp != "" {
		merged.PublicIp = h.PublicIp
	}
	if h.PrivateIp != "" {
		merged.PrivateIp = h.PrivateIp
	}
	if h.InstanceType != "" {
		merged.InstanceType = h.InstanceType
	}
	if h.InstanceResourceType != "" {
		merged.InstanceResourceType = h.InstanceResourceType
	}
	if h.SpotPrice != "" {
		merged.SpotPrice = h.SpotPrice
	}
	if h.SpotType != "" {
		merged.SpotType = h.SpotType
	}
	if merged.Components, err = h.Components.Merge(parent.Components); err != nil {
		return nil, fmt.Errorf("error merging components (%w)", err)
	}

	return merged, nil
}

func (c Components) Merge(parent Components) (Components, error) {
	merged := make(Components)
	if err := copier.Copy(&merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	for k, v := range c {
		if pv, found := merged[k]; found {
			if mergedComponent, err := v.Merge(pv); err == nil {
				merged[k] = mergedComponent
			} else {
				return nil, fmt.Errorf("error merging component (%w)", err)
			}
		} else {
			merged[k] = v
		}
	}

	return merged, nil
}

func (c *Component) Merge(parent *Component) (*Component, error) {
	merged := &Component{}
	if err := copier.Copy(merged, parent); err != nil {
		return nil, fmt.Errorf("error copying parent (%w)", err)
	}

	var err error
	if merged.Scope, err = c.Scope.Merge(parent.Scope); err != nil {
		return nil, fmt.Errorf("error merging component scope (%w)", err)
	}

	if c.ScriptSrc != "" {
		merged.ScriptSrc = c.ScriptSrc
	}
	if c.ScriptName != "" {
		merged.ScriptName = c.ScriptName
	}
	if c.ConfigSrc != "" {
		merged.ConfigSrc = c.ConfigSrc
	}
	if c.ConfigName != "" {
		merged.ConfigName = c.ConfigName
	}
	if c.BinaryName != "" {
		merged.BinaryName = c.BinaryName
	}
	if c.PublicIdentity != "" {
		merged.PublicIdentity = c.PublicIdentity
	}
	if c.PrivateIdentity != "" {
		merged.PrivateIdentity = c.PrivateIdentity
	}

	return merged, nil
}
