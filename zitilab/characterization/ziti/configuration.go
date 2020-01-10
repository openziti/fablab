package zitilab_characterization_ziti

import (
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/kernel/runlevel/1_configuration/config"
	"github.com/netfoundry/fablab/kernel/runlevel/1_configuration/pki"
)

func newConfigurationFactory() model.Factory {
	return &configurationFactory{}
}

func (f *configurationFactory) Build(m *model.Model) error {
	m.Configuration = model.ConfigurationBinders{
		func(m *model.Model) model.ConfigurationStage { return pki.Group(pki.Fabric(), pki.DotZiti()) },
		func(m *model.Model) model.ConfigurationStage { return config.Component() },
		func(m *model.Model) model.ConfigurationStage {
			configs := []config.StaticConfig{
				{Src: "loop/10-ambient.loop2.yml", Name: "10-ambient.loop2.yml"},
				{Src: "loop/4k-chatter.loop2.yml", Name: "4k-chatter.loop2.yml"},
				{Src: "remote_identities.yml", Name: "remote_identities.yml"},
			}
			return config.Static(configs)
		},
	}
	return nil
}

type configurationFactory struct{}
