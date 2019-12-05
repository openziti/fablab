package cli

import (
	"github.com/netfoundry/fablab/kernel"
	"github.com/netfoundry/fablab/zitilab"
	"bytes"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

func Fabric(args ...string) kernel.Action {
	return &fabric{
		args: args,
	}
}

func (a *fabric) Execute(m *kernel.Model) error {
	if !m.IsBound() {
		return errors.New("model not bound")
	}

	allArgs := append(a.args, "-i", "fablab")
	cli := exec.Command(zitilab.ZitiFabricCli(), allArgs...)
	var cliOut bytes.Buffer
	cli.Stdout = &cliOut
	var cliErr bytes.Buffer
	cli.Stderr = &cliErr
	logrus.Infof("%v", cli.Args)
	err := cli.Run()
	out := fmt.Sprintf("out:[%s], err:[%s]", strings.Trim(cliOut.String(), " \t\r\n"), strings.Trim(cliErr.String(), " \t\r\n"))
	logrus.Info(out)
	if err != nil {
		return err
	}
	return nil
}

type fabric struct {
	args []string
}