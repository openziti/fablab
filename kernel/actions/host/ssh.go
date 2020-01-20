package host

import "github.com/netfoundry/fablab/kernel/internal"

func RemoteShell(user, host, keyPath string) error {
	factory := internal.NewSshConfigFactoryImpl(user, host)
	factory.KeyPath = keyPath

	return internal.RemoteShell(factory)
}
