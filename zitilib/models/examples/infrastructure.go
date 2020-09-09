package zitilib_examples

import (
	"fmt"
	aws_ssh_keys0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/aws_ssh_key"
	semaphore0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/semaphore"
	terraform0 "github.com/openziti/fablab/kernel/fablib/runlevel/0_infrastructure/terraform"
	aws_ssh_keys6 "github.com/openziti/fablab/kernel/fablib/runlevel/6_disposal/aws_ssh_key"
	terraform6 "github.com/openziti/fablab/kernel/fablib/runlevel/6_disposal/terraform"
	"github.com/openziti/fablab/kernel/model"
	"time"
)

func newInfrastructureFactory() model.Factory {
	return &infrastructureFactory{}
}

func (self *infrastructureFactory) Build(m *model.Model) error {
	if err := self.buildInfrastructure(m); err != nil {
		return fmt.Errorf("error building infrastructure bindings (%w)", err)
	}
	if err := self.buildDisposal(m); err != nil {
		return fmt.Errorf("error building disposal bindings (%w)", err)
	}
	return nil
}

func (_ *infrastructureFactory) buildInfrastructure(m *model.Model) error {
	m.Infrastructure = model.InfrastructureStages{
		aws_ssh_keys0.Express(),
		terraform0.Express(),
		semaphore0.Restart(90 * time.Second),
	}
	return nil
}

func (_ *infrastructureFactory) buildDisposal(m *model.Model) error {
	m.Disposal = model.DisposalStages{
		terraform6.Dispose(),
		aws_ssh_keys6.Dispose(),
	}
	return nil
}

type infrastructureFactory struct{}
