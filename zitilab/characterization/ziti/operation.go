package zitilab_characterization_ziti

import (
	"github.com/netfoundry/fablab/kernel/model"
	operation "github.com/netfoundry/fablab/kernel/runlevel/5_operation"
	"time"
)

func newOperationFactory() model.Factory {
	return &operationFactory{}
}

func (f *operationFactory) Build(m *model.Model) error {
	c := make(chan struct{})
	m.Operation = model.OperatingBinders{
		func(m *model.Model) model.OperatingStage { return operation.Mesh(c) },
		func(m *model.Model) model.OperatingStage { return operation.Metrics(c) },
		func(m *model.Model) model.OperatingStage {
			minutes, found := m.GetVariable("sample_minutes")
			if !found {
				minutes = 1
			}
			sampleDuration := time.Duration(minutes.(int)) * time.Minute
			return operation.Iperf(int(sampleDuration.Seconds()))
		},
		func(m *model.Model) model.OperatingStage { return operation.Closer(c) },
		func(m *model.Model) model.OperatingStage { return operation.Persist() },
	}
	return nil
}

type operationFactory struct{}