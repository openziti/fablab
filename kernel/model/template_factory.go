package model

import (
	"bytes"
	"github.com/openziti/foundation/util/errorz"
	"github.com/pkg/errors"
	"html/template"
)

type TemplatingStrategy interface {
	IsTemplated(entity Entity) bool
	GetEntityCount(entity Entity) int
}

type TemplatingFactory struct {
	Strategy TemplatingStrategy
}

func (factory *TemplatingFactory) Build(m *Model) error {
	if err := factory.ProcessRegions(m); err != nil {
		return err
	}

	if err := factory.ProcessHosts(m); err != nil {
		return err
	}

	return factory.ProcessComponents(m)
}

func (factory *TemplatingFactory) ProcessRegions(m *Model) error {
	var templatedRegions []*Region

	for key, region := range m.Regions {
		if factory.Strategy.IsTemplated(region) {
			delete(m.Regions, key)
			templatedRegions = append(templatedRegions, region)
		}
	}

	for _, region := range templatedRegions {
		scaleFactor := factory.Strategy.GetEntityCount(region)
		for idx := 0; idx < scaleFactor; idx++ {
			cloned := region.CloneRegion(idx)

			templater := &Templater{data: cloned}
			newKey := templater.Templatize(region.Id)
			cloned.init(newKey, m)
			cloned.Templatize(templater)

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

func (factory *TemplatingFactory) ProcessHosts(m *Model) error {
	var templatedHosts []*Host

	for _, region := range m.Regions {
		for key, host := range region.Hosts {
			if factory.Strategy.IsTemplated(host) {
				delete(region.Hosts, key)
				templatedHosts = append(templatedHosts, host)
			}
		}
	}

	for _, host := range templatedHosts {
		scaleFactor := factory.Strategy.GetEntityCount(host)
		for idx := 0; idx < scaleFactor; idx++ {
			cloned := host.CloneHost(idx)

			templater := &Templater{data: cloned}
			newKey := templater.Templatize(host.Id)
			cloned.init(newKey, host.Region)
			cloned.Templatize(templater)

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

func (factory *TemplatingFactory) ProcessComponents(m *Model) error {
	var templatedComponents []*Component

	for _, region := range m.Regions {
		for _, host := range region.Hosts {
			for key, component := range host.Components {
				if factory.Strategy.IsTemplated(component) {
					delete(host.Components, key)
					templatedComponents = append(templatedComponents, component)
				}
			}
		}
	}

	for _, component := range templatedComponents {
		scaleFactor := factory.Strategy.GetEntityCount(component)
		for idx := 0; idx < scaleFactor; idx++ {
			cloned := component.CloneComponent(idx)

			templater := &Templater{data: cloned}
			newKey := templater.Templatize(component.Id)
			cloned.init(newKey, component.Host)
			cloned.Templatize(templater)

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

type Templater struct {
	errorz.ErrorHolderImpl
	data interface{}
}

func (t *Templater) Templatize(val string) string {
	tmpl := template.New("model")
	tmpl, err := tmpl.Parse(val)
	if t.SetError(err) {
		return val
	}
	buf := bytes.NewBuffer(nil)
	if t.SetError(tmpl.Execute(buf, t.data)) {
		return val
	}
	return string(buf.Bytes())
}
