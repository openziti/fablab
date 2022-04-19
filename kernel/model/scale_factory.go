package model

import (
	"github.com/pkg/errors"
)

type ScaleStrategy interface {
	IsScaled(entity Entity) bool
	GetEntityCount(entity Entity) uint32
}

type ScaleEntityFactory interface {
	CreateScaledRegion(source *Region, scaleIndex uint32) (*Region, error)
	CreateScaledHost(source *Host, scaleIndex uint32) (*Host, error)
	CreateScaledComponent(source *Component, scaleIndex uint32) (*Component, error)
}

func NewScaleFactory(strategy ScaleStrategy, factory ScaleEntityFactory) *ScaleFactory {
	return &ScaleFactory{
		Strategy:      strategy,
		EntityFactory: factory,
	}
}

func NewScaleFactoryWithDefaultEntityFactory(strategy ScaleStrategy) *ScaleFactory {
	return NewScaleFactory(strategy, DefaultScaleEntityFactory{})
}

type ScaleFactory struct {
	Strategy      ScaleStrategy
	EntityFactory ScaleEntityFactory
}

func (factory *ScaleFactory) Build(m *Model) error {
	defer m.Accept(func(entity Entity) {
		delete(entity.GetScope().Defaults, "__scaled__")
	})

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

	m.RangeSortedRegions(func(id string, region *Region) {
		if factory.Strategy.IsScaled(region) {
			m.RemoveRegion(region)
			scaledRegions = append(scaledRegions, region)
		}
	})

	for _, region := range scaledRegions {
		scaleFactor := factory.Strategy.GetEntityCount(region)
		for idx := uint32(0); idx < scaleFactor; idx++ {
			cloned, err := factory.EntityFactory.CreateScaledRegion(region, idx)
			if err != nil {
				return err
			}

			if _, found := m.Regions[cloned.Id]; found {
				return errors.Errorf("region with id %v already exists. Either set scale to 1 instead of %v or change the id", cloned.Id, scaleFactor)
			}

			m.Regions[cloned.Id] = cloned
			factory.markScaled(cloned)
		}
	}

	return nil
}

func (factory *ScaleFactory) isParentScaled(entity Entity) bool {
	_, found := entity.GetParentEntity().GetScope().Defaults["__scaled__"]
	return found
}

func (factory *ScaleFactory) markScaled(entity Entity) {
	entity.GetScope().Defaults["__scaled__"] = struct{}{}
}

func (factory *ScaleFactory) ProcessHosts(m *Model) error {
	var scaledHosts []*Host

	m.RangeSortedRegions(func(id string, region *Region) {
		region.RangeSortedHosts(func(id string, host *Host) {
			if factory.isParentScaled(host) || factory.Strategy.IsScaled(host) {
				region.RemoveHost(host)
				scaledHosts = append(scaledHosts, host)
			}
		})
	})

	for _, host := range scaledHosts {
		var scaleFactor uint32 = 1
		if factory.Strategy.IsScaled(host) {
			scaleFactor = factory.Strategy.GetEntityCount(host)
		}
		for idx := uint32(0); idx < scaleFactor; idx++ {
			cloned, err := factory.EntityFactory.CreateScaledHost(host, idx)
			if err != nil {
				return err
			}

			if _, found := host.Region.Hosts[cloned.Id]; found {
				return errors.Errorf("host with id %v > %v already exists. Either set scale to 1 instead of %v or change the id",
					host.Region.Id, cloned.Id, scaleFactor)
			}
			host.Region.Hosts[cloned.Id] = cloned
			factory.markScaled(cloned)
		}
	}

	return nil
}

func (factory *ScaleFactory) ProcessComponents(m *Model) error {
	var scaledComponents []*Component

	m.RangeSortedRegions(func(id string, region *Region) {
		region.RangeSortedHosts(func(id string, host *Host) {
			host.RangeSortedComponents(func(id string, component *Component) {
				if factory.isParentScaled(component) || factory.Strategy.IsScaled(component) {
					host.RemoveComponent(component)
					scaledComponents = append(scaledComponents, component)
				}
			})
		})
	})

	for _, component := range scaledComponents {
		var scaleFactor uint32 = 1
		if factory.Strategy.IsScaled(component) {
			scaleFactor = factory.Strategy.GetEntityCount(component)
		}
		for idx := uint32(0); idx < scaleFactor; idx++ {
			cloned, err := factory.EntityFactory.CreateScaledComponent(component, idx)
			if err != nil {
				return err
			}

			if _, found := component.Host.Components[cloned.Id]; found {
				return errors.Errorf("component with id %v > %v > %v already exists. Either set scale to 1 instead of %v or change the id",
					component.Host.Region.Id, component.Host.Id, cloned.Id, scaleFactor)
			}
			component.Host.Components[cloned.Id] = cloned
			factory.markScaled(cloned)
		}
	}

	return nil
}

type DefaultScaleEntityFactory struct{}

func (self DefaultScaleEntityFactory) CreateScaledRegion(source *Region, scaleIndex uint32) (*Region, error) {
	cloned := source.CloneRegion(scaleIndex)

	templater := &Templater{data: cloned}
	newKey := templater.TemplatizeString(source.Id)
	cloned.init(newKey, source.GetModel())
	templater.TemplatizeRegion(cloned)

	if templater.HasError() {
		return nil, templater.GetError()
	}

	return cloned, nil
}

func (self DefaultScaleEntityFactory) CreateScaledHost(source *Host, scaleIndex uint32) (*Host, error) {
	cloned := source.CloneHost(scaleIndex)

	templater := &Templater{data: cloned}
	newKey := templater.TemplatizeString(source.Id)
	cloned.init(newKey, source.GetRegion())
	templater.TemplatizeHost(cloned)

	if templater.HasError() {
		return nil, templater.GetError()
	}

	return cloned, nil
}

func (self DefaultScaleEntityFactory) CreateScaledComponent(source *Component, scaleIndex uint32) (*Component, error) {
	cloned := source.CloneComponent(scaleIndex)

	templater := &Templater{data: cloned}
	newKey := templater.TemplatizeString(source.Id)
	cloned.init(newKey, source.GetHost())
	templater.TemplatizeComponent(cloned)

	if templater.HasError() {
		return nil, templater.GetError()
	}

	return cloned, nil
}
