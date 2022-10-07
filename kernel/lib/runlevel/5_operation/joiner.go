/*
	Copyright 2020 NetFoundry Inc.

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

package operation

import (
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Joiner(joiners []chan struct{}) model.OperatingStage {
	return &joiner{
		joiners: joiners,
	}
}

func (j *joiner) Operate(model.Run) error {
	logrus.Debugf("will join with [%d] joiners", len(j.joiners))
	count := 0
	for _, joiner := range j.joiners {
		<-joiner
		logrus.Debugf("joined with joiner [%d]", count)
		count++
	}
	logrus.Infof("joined with [%d] joiners", len(j.joiners))
	return nil
}

type joiner struct {
	joiners []chan struct{}
}
