package operation

import (
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Joiner(joiners []chan struct{}) model.OperatingStage {
	return &joiner{
		joiners: joiners,
	}
}

func (j *joiner) Operate(m *model.Model, _ string) error {
	logrus.Debugf("will join with [%d] joiners", len(j.joiners))
	count := 0
	for _, joiner := range j.joiners {
		<-joiner
		logrus.Debugf("joined with joiner [%d]", count)
		count++
	}
	logrus.Infof("joined with [%d] joiners", len(j.joiners))
	return nil
}

type joiner struct {
	joiners []chan struct{}
}
