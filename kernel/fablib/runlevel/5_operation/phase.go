package operation

import (
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func NewPhase() Phase {
	return &phaseImpl{
		closer:  make(chan struct{}),
		joiners: nil,
	}
}

type Phase interface {
	model.OperatingStage
	GetCloser() <-chan struct{}
	AddJoiner() chan struct{}
}

func (phase *phaseImpl) Operate(model.Run) error {
	logrus.Debugf("waiting for [%d] tasks to complete", len(phase.joiners))
	count := 0
	for _, joiner := range phase.joiners {
		<-joiner
		logrus.Debugf("task [%d] completed", count)
		count++
	}
	logrus.Infof("[%d] tasks completed", len(phase.joiners))

	logrus.Info("phase complete")
	close(phase.closer)

	return nil
}

func (phase *phaseImpl) GetCloser() <-chan struct{} {
	return phase.closer
}

func (phase *phaseImpl) AddJoiner() chan struct{} {
	joinerChan := make(chan struct{})
	phase.joiners = append(phase.joiners, joinerChan)
	return joinerChan
}

type phaseImpl struct {
	closer  chan struct{}
	joiners []chan struct{}
}
