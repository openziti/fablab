package fablib

import (
	"github.com/natefinch/npipe"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"time"
)

func sshAuthMethodAgent() ssh.AuthMethod {
	if sshAgent, err := npipe.DialTimeout(`\\.\pipe\openssh-ssh-agent`, 1*time.Second); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	} else {
		logrus.WithError(err).Warn("could not connect to ssh-agent pipe")
	}
	return nil
}
