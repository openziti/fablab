package __operation

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func LoopListener(host *model.Host, joiner chan struct{}) model.OperatingStage {
	return &loopListener{
		host:   host,
		joiner: joiner,
	}
}

func (self *loopListener) Operate(m *model.Model, run string) error {
	ssh := fablib.NewSshConfigFactoryImpl(m, self.host.PublicIp)
	if err := fablib.RemoteKill(ssh, "ziti-fabric-test loop2 listener"); err != nil {
		return fmt.Errorf("error killing loop2 listeners (%w)", err)
	}

	go self.run(m, run)
	return nil
}

func (self *loopListener) run(m *model.Model, run string) {
	defer func() {
		if self.joiner != nil {
			close(self.joiner)
			logrus.Debugf("closed joiner")
		}
	}()

	ssh := fablib.NewSshConfigFactoryImpl(m, self.host.PublicIp)

	logFile := fmt.Sprintf("/home/%s/logs/loop2-listener-%s.log", ssh.User(), run)
	listenerCmd := fmt.Sprintf("/home/%s/fablab/bin/ziti-fabric-test loop2 listener -b tcp:0.0.0.0:8171 >> %s 2>&1", ssh.User(), logFile)
	if output, err := fablib.RemoteExec(ssh, listenerCmd); err != nil {
		logrus.Error("error starting loop listener [%s] (%w)", output, err)
	}
}

type loopListener struct {
	host   *model.Host
	joiner chan struct{}
}
