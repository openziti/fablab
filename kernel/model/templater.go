package model

import (
	"bytes"
	"github.com/openziti/foundation/v2/errorz"
	"text/template"
)

type Templater struct {
	errorz.ErrorHolderImpl
	data interface{}
}

func (self *Templater) TemplatizeString(val string) string {
	tmpl := template.New("model")
	tmpl, err := tmpl.Parse(val)
	if self.SetError(err) {
		return val
	}
	buf := bytes.NewBuffer(nil)
	if self.SetError(tmpl.Execute(buf, self.data)) {
		return val
	}
	return buf.String()
}

func (self *Templater) TemplatizeScope(scope *Scope) {
	scope.Defaults.ForEach(func(k string, v interface{}) (bool, interface{}) {
		if strVal, ok := v.(string); ok {
			return true, self.TemplatizeString(strVal)
		}
		return false, nil
	})

	var newTags Tags
	for _, tag := range scope.Tags {
		newTags = append(newTags, self.TemplatizeString(tag))
	}
	scope.Tags = newTags
}

func (self *Templater) TemplatizeRegion(region *Region) {
	self.TemplatizeScope(&region.Scope)
	region.Region = self.TemplatizeString(region.Region)
	region.Site = self.TemplatizeString(region.Site)
}

func (self *Templater) TemplatizeHost(host *Host) {
	self.TemplatizeScope(&host.Scope)
	host.PublicIp = self.TemplatizeString(host.PublicIp)
	host.PrivateIp = self.TemplatizeString(host.PrivateIp)
	host.InstanceType = self.TemplatizeString(host.InstanceType)
	host.InstanceResourceType = self.TemplatizeString(host.InstanceResourceType)
	host.SpotPrice = self.TemplatizeString(host.SpotPrice)
	host.SpotType = self.TemplatizeString(host.SpotType)
}

func (self *Templater) TemplatizeComponent(component *Component) {
	self.TemplatizeScope(&component.Scope)
}
