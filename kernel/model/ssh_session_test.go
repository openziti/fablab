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

package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSshSessionLimit(t *testing.T) {
	require.Equal(t, defaultSshSessionLimit, sshSessionLimit(0), "unset uses the default")
	require.Equal(t, defaultSshSessionLimit, sshSessionLimit(-1), "negative uses the default")
	require.Equal(t, 1, sshSessionLimit(1))
	require.Equal(t, 25, sshSessionLimit(25))
}
