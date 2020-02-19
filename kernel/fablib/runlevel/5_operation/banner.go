package operation

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
)

func Banner(msg string) model.OperatingStage {
	return &banner{msg: msg}
}

func (b *banner) Operate(_ *model.Model, _ string) error {
	fablib.Figlet(b.msg)
	fmt.Println()
	return nil
}

type banner struct{
	msg string
}
