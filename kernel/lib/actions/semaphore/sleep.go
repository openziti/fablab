/*
	Copyright 2019 NetFoundry Inc.

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

package semaphore

import (
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"time"
)

func Sleep(duration time.Duration) model.Action {
	return &sleep{duration: duration}
}

func (sleep *sleep) Execute(_ model.Run) error {
	logrus.Infof("sleeping for [%s]", sleep.duration)
	time.Sleep(sleep.duration)
	return nil
}

type sleep struct {
	duration time.Duration
}
