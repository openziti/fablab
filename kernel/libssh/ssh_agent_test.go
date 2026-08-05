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

package libssh

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestSshAgentAuthMethodIsShared(t *testing.T) {
	previousFactory := sshAgentAuthMethodFactory
	t.Cleanup(func() {
		sshAgentAuthMethodFactory = previousFactory
		sharedSshAgentAuth = sshAgentAuthCache{}
	})

	createCount := 0
	sshAgentAuthMethodFactory = func() ssh.AuthMethod {
		createCount++
		return ssh.Password("test")
	}
	sharedSshAgentAuth = sshAgentAuthCache{}

	first := sshAuthMethodAgent()
	require.NotNil(t, first)
	for range 100 {
		require.NotNil(t, sshAuthMethodAgent())
	}

	require.Equal(t, 1, createCount)
}
