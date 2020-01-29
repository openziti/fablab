package host

import (
	"github.com/netfoundry/fablab/kernel/internal"
	"github.com/netfoundry/fablab/kernel/model"
)

func RemoteShell(m *model.Model, host string) error {
	factory := internal.NewSshConfigFactoryImpl(m, host)

	return internal.RemoteShell(factory)
}
