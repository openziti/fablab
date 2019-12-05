package pki

import (
	"github.com/netfoundry/fablab/kernel"
	"fmt"
	"github.com/sirupsen/logrus"
)

func Group(stages ...kernel.ConfigurationStage) kernel.ConfigurationStage {
	return &group{stages: stages}
}

func (group *group) Configure(m *kernel.Model) error {
	if existing, err := hasExisitingPki(); err == nil {
		if existing {
			logrus.Infof("skipping configuration. existing pki system at [%s]", kernel.PkiBuild())
			return nil
		}
	} else {
		return fmt.Errorf("error checking pki existence at [%s] (%s)", kernel.PkiBuild(), err)
	}

	for _, stage := range group.stages {
		if err := stage.Configure(m); err != nil {
			return fmt.Errorf("error running configuration stage (%w)", err)
		}
	}

	return nil
}

type group struct {
	stages []kernel.ConfigurationStage
}
