package model

import (
	"github.com/pkg/errors"
)

type ScaleStrategy interface {
	IsScaled(entity Entity) bool
	GetEntityCount(entity Entity) int
}

func NewScaleFactory(strategy ScaleStrategy) *ScaleFactory {
	return &ScaleFactory{
		Strategy: strategy,
	}
}

type ScaleFactory struct {
	Strategy ScaleStrategy
}

func (factory *ScaleFactory) Build(m *Model) error {
	if err := factory.ProcessRegions(m); err != nil {
		return err
	}

	if err := factory.ProcessHosts(m); err != nil {
		return err
	}

	return factory.ProcessComponents(m)
}

func (factory *ScaleFactory) ProcessRegions(m *Model) error {
	var scaledRegions []*Region

	for key, region := range m.Regions {
		if factory.Strategy.IsScaled(region) {
			delete(m.Regions, key)
			scaledRegions = append(scaledRegions, region)
		}
	}

	for _, region := range scaledRegions {
		scaleFactor := factory.Strategy.GetEntityCount(region)
		for idx := 0; idx < scaleFactor; idx++ {
			cloned := region.CloneRegion(idx)

			templater := &Templater{data: cloned}
			newKey := templater.TemplatizeString(region.Id)
			cloned.init(newKey, m)
			templater.TemplatizeRegion(cloned)

			if _, found := m.Regions[newKey]; found {
				return errors.Errorf("region with id %v already exists. Either set scale to 1 instead of %v or templatize id", newKey, scaleFactor)
			}

			m.Regions[newKey] = cloned

			if templater.HasError() {
				return templater.GetError()
			}
		}
	}

	return nil
}

func (factory *ScaleFactory) ProcessHosts(m *Model) error {
	var scaledHosts []*Host

	for _, region := range m.Regions {
		for key, host := range region.Hosts {
			if factory.Strategy.IsScaled(host) {
				delete(region.Hosts, key)
				scaledHosts = append(scaledHosts, host)
			}
		}
	}

	for _, host := range scaledHosts {
		scaleFactor := factory.Strategy.GetEntityCount(host)
		for idx := 0; idx < scaleFactor; idx++ {
			cloned := host.CloneHost(idx)

			templater := &Templater{data: cloned}
			newKey := templater.TemplatizeString(host.Id)
			cloned.init(newKey, host.Region)
			templater.TemplatizeHost(cloned)

			if _, found := host.Region.Hosts[newKey]; found {
				return errors.Errorf("host with id %v > %v already exists. Either set scale to 1 instead of %v or templatize id",
					host.Region.Id, newKey, scaleFactor)
			}
			host.Region.Hosts[newKey] = cloned

			if templater.HasError() {
				return templater.GetError()
			}
		}
	}

	return nil
}

func (factory *ScaleFactory) ProcessComponents(m *Model) error {
	var scaledComponents []*Component

	for _, region := range m.Regions {
		for _, host := range region.Hosts {
			for key, component := range host.Components {
				if factory.Strategy.IsScaled(component) {
					delete(host.Components, key)
					scaledComponents = append(scaledComponents, component)
				}
			}
		}
	}

	for _, component := range scaledComponents {
		scaleFactor := factory.Strategy.GetEntityCount(component)
		for idx := 0; idx < scaleFactor; idx++ {
			cloned := component.CloneComponent(idx)

			templater := &Templater{data: cloned}
			newKey := templater.TemplatizeString(component.Id)
			cloned.init(newKey, component.Host)
			templater.TemplatizeComponent(cloned)

			if _, found := component.Host.Components[newKey]; found {
				return errors.Errorf("component with id %v > %v > %v already exists. Either set scale to 1 instead of %v or templatize id",
					component.Host.Region.Id, component.Host.Id, newKey, scaleFactor)
			}
			component.Host.Components[newKey] = cloned

			if templater.HasError() {
				return templater.GetError()
			}
		}
	}

	return nil
}
