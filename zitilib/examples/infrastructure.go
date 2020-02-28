package zitilib_examples

import (
	"fmt"
	semaphore0 "github.com/netfoundry/fablab/kernel/fablib/runlevel/0_infrastructure/semaphore"
	terraform0 "github.com/netfoundry/fablab/kernel/fablib/runlevel/0_infrastructure/terraform"
	terraform6 "github.com/netfoundry/fablab/kernel/fablib/runlevel/6_disposal/terraform"
	"github.com/netfoundry/fablab/kernel/model"
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
	m.Infrastructure = model.InfrastructureBinders{
		func(m *model.Model) model.InfrastructureStage { return terraform0.Express() },
		func(m *model.Model) model.InfrastructureStage { return semaphore0.Restart(90 * time.Second) },
	}
	return nil
}

func (_ *infrastructureFactory) buildDisposal(m *model.Model) error {
	m.Disposal = model.DisposalBinders{
		func(m *model.Model) model.DisposalStage { return terraform6.Dispose() },
	}
}

type infrastructureFactory struct{}