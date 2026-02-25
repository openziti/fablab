/*
	(c) Copyright NetFoundry Inc. Inc.

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

package tui

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/sirupsen/logrus"
	"sync/atomic"
)

const PaneField = "tui.pane"

const (
	PaneActions    = "actions"
	PaneValidation = "validation"
)

var active atomic.Bool

// IsActive returns true when the TUI mode is active.
func IsActive() bool {
	return active.Load()
}

// ActionsLogger returns a logger whose output routes to the actions pane when TUI is active.
// When TUI is not active, returns a plain logger without the routing field.
func ActionsLogger() *logrus.Entry {
	if active.Load() {
		return pfxlog.Logger().WithField(PaneField, PaneActions)
	}
	return pfxlog.Logger().WithFields(nil)
}

// ValidationLogger returns a logger whose output routes to the validation pane when TUI is active.
// When TUI is not active, returns a plain logger without the routing field.
func ValidationLogger() *logrus.Entry {
	if active.Load() {
		return pfxlog.Logger().WithField(PaneField, PaneValidation)
	}
	return pfxlog.Logger().WithFields(nil)
}
