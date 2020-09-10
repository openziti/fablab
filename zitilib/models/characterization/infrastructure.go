/*
	Copyright 2020 NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package zitilib_characterization

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

func (f *infrastructureFactory) Build(m *model.Model) error {
	if err := f.buildInfrastructure(m); err != nil {
		return fmt.Errorf("error building infrastructure bindings (%w)", err)
	}
	if err := f.buildDisposal(m); err != nil {
		return fmt.Errorf("error building disposal bindings (%w)", err)
	}
	return nil
}

func (f *infrastructureFactory) buildInfrastructure(m *model.Model) error {
	m.Infrastructure = model.InfrastructureStages{
		aws_ssh_keys0.Express(),
		terraform0.Express(),
		semaphore0.Restart(90 * time.Second),
	}
	return nil
}

func (f *infrastructureFactory) buildDisposal(m *model.Model) error {
	m.Disposal = model.DisposalStages{
		terraform6.Dispose(),
		aws_ssh_keys6.Dispose(),
	}
	return nil
}

type infrastructureFactory struct{}
