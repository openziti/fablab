package zitilib_runlevel_5_operation

import (
	"fmt"
	"github.com/openziti/fablab/kernel/fablib"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"strings"
)

func LoopListener(host *model.Host, joiner chan struct{}, bindAddress string, extraArgs ...string) model.OperatingStage {
	return &loopListener{
		host:        host,
		joiner:      joiner,
		bindAddress: bindAddress,
		extraArgs:   extraArgs,
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
	listenerCmd := fmt.Sprintf("/home/%s/fablab/bin/ziti-fabric-test loop2 listener -b %v %v >> %s 2>&1",
		ssh.User(), self.bindAddress, strings.Join(self.extraArgs, " "), logFile)
	if output, err := fablib.RemoteExec(ssh, listenerCmd); err != nil {
		logrus.Errorf("error starting loop listener [%s] (%v)", output, err)
	}
}

type loopListener struct {
	host        *model.Host
	joiner      chan struct{}
	bindAddress string
	extraArgs   []string
}
