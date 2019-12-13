package zitilab_bootstrap

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"os"
)

func (bootstrap *Bootstrap) Bootstrap(m *model.Model) error {
	zitiRoot = os.Getenv("ZITI_ROOT")
	if zitiRoot == "" {
		return fmt.Errorf("please set 'ZITI_ROOT'")
	}
	if fi, err := os.Stat(zitiRoot); err == nil {
		if !fi.IsDir() {
			return fmt.Errorf("invalid 'ZITI_ROOT' (!directory)")
		}
		logrus.Debugf("ZITI_ROOT = [%s]", zitiRoot)
	} else {
		return fmt.Errorf("non-existent 'ZITI_ROOT'")
	}
	return nil
}

type Bootstrap struct{}
