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

	"github.com/stretchr/testify/assert"
)

func TestEnsureFablabChainsCmd(t *testing.T) {
	cmd := ensureFablabChainsCmd()
	assert.Contains(t, cmd, "sudo iptables -N FABLAB_INPUT 2>/dev/null || true")
	assert.Contains(t, cmd, "sudo iptables -N FABLAB_OUTPUT 2>/dev/null || true")
	assert.Contains(t, cmd, "sudo iptables -C INPUT -j FABLAB_INPUT 2>/dev/null || sudo iptables -I INPUT -j FABLAB_INPUT")
	assert.Contains(t, cmd, "sudo iptables -C OUTPUT -j FABLAB_OUTPUT 2>/dev/null || sudo iptables -I OUTPUT -j FABLAB_OUTPUT")
}

func TestBlockIncomingCmd(t *testing.T) {
	assert.Equal(t, "sudo iptables -A FABLAB_INPUT -p tcp --dport 8080 -j DROP", blockIncomingCmd(8080))
	assert.Equal(t, "sudo iptables -A FABLAB_INPUT -p tcp --dport 443 -j DROP", blockIncomingCmd(443))
}

func TestUnblockIncomingCmd(t *testing.T) {
	assert.Equal(t, "sudo iptables -D FABLAB_INPUT -p tcp --dport 8080 -j DROP", unblockIncomingCmd(8080))
}

func TestBlockOutgoingCmd(t *testing.T) {
	assert.Equal(t, "sudo iptables -A FABLAB_OUTPUT -p tcp -d 10.0.0.1 --dport 6262 -j DROP", blockOutgoingCmd("10.0.0.1", 6262))
}

func TestUnblockOutgoingCmd(t *testing.T) {
	assert.Equal(t, "sudo iptables -D FABLAB_OUTPUT -p tcp -d 10.0.0.1 --dport 6262 -j DROP", unblockOutgoingCmd("10.0.0.1", 6262))
}

func TestUnblockAllCmd(t *testing.T) {
	cmd := unblockAllCmd()
	assert.Contains(t, cmd, "sudo iptables -F FABLAB_INPUT")
	assert.Contains(t, cmd, "sudo iptables -F FABLAB_OUTPUT")
}

func TestKillIncomingCmd(t *testing.T) {
	assert.Equal(t, "sudo ss -K -t state established sport = :8080", killIncomingCmd(8080))
}

func TestKillOutgoingCmd(t *testing.T) {
	assert.Equal(t, "sudo ss -K -t state established dst 10.0.0.1:6262", killOutgoingCmd("10.0.0.1", 6262))
}

func TestBlockUnblockSymmetry(t *testing.T) {
	// Block and unblock commands should differ only by -A vs -D
	block := blockIncomingCmd(8080)
	unblock := unblockIncomingCmd(8080)
	assert.Contains(t, block, "-A FABLAB_INPUT")
	assert.Contains(t, unblock, "-D FABLAB_INPUT")

	blockOut := blockOutgoingCmd("10.0.0.1", 6262)
	unblockOut := unblockOutgoingCmd("10.0.0.1", 6262)
	assert.Contains(t, blockOut, "-A FABLAB_OUTPUT")
	assert.Contains(t, unblockOut, "-D FABLAB_OUTPUT")
}
