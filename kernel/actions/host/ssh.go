package host

import "github.com/netfoundry/fablab/kernel/internal"

func RemoteShell(user, host string) error {
	return internal.RemoteShell(user, host)
}
