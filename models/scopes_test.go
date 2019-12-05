package models

import (
	"github.com/netfoundry/fablab/kernel"
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestIterateScopes(t *testing.T) {
	diamondback.IterateScopes(func(i interface{}, path ...string) {
		if m, ok := i.(*kernel.Model); ok {
			logrus.Infof("model, tags = %v", m.Tags)
		} else if r, ok := i.(*kernel.Region); ok {
			logrus.Infof("region %v, tags = %v", path, r.Tags)
		} else if h, ok := i.(*kernel.Host); ok {
			logrus.Infof("host %v, tags = %v", path, h.Tags)
		} else if c, ok := i.(*kernel.Component); ok {
			logrus.Infof("component %v, tags = %v", path, c.Tags)
		} else {
			logrus.Infof("%v, s = %p, %s", path, i, reflect.TypeOf(i))
		}
	})
}