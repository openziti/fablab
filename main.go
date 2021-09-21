package fablab

import (
	"github.com/openziti/fablab/cmd/fablab/subcmd"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func InitModel(m *model.Model) {
	model.InitModel(m)
}

func Run() {
	if err := subcmd.RootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatal("failure")
	}
}
