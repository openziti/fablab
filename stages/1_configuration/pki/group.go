package pki

import (
	"fmt"
	"github.com/netfoundry/fablab/model"
	"github.com/sirupsen/logrus"
)

func Group(stages ...model.ConfigurationStage) model.ConfigurationStage {
	return &group{stages: stages}
}

func (group *group) Configure(m *model.Model) error {
	if existing, err := hasExisitingPki(); err == nil {
		if existing {
			logrus.Infof("skipping configuration. existing pki system at [%s]", model.PkiBuild())
			return nil
		}
	} else {
		return fmt.Errorf("error checking pki existence at [%s] (%s)", model.PkiBuild(), err)
	}

	for _, stage := range group.stages {
		if err := stage.Configure(m); err != nil {
			return fmt.Errorf("error running configuration stage (%w)", err)
		}
	}

	return nil
}

type group struct {
	stages []model.ConfigurationStage
}
