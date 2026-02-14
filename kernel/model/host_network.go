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
	"fmt"

	"github.com/sirupsen/logrus"
)

// command construction helpers (unexported, testable)

func ensureFablabChainsCmd() string {
	return "sudo iptables -N FABLAB_INPUT 2>/dev/null || true" +
		" && sudo iptables -N FABLAB_OUTPUT 2>/dev/null || true" +
		" && (sudo iptables -C INPUT -j FABLAB_INPUT 2>/dev/null || sudo iptables -I INPUT -j FABLAB_INPUT)" +
		" && (sudo iptables -C OUTPUT -j FABLAB_OUTPUT 2>/dev/null || sudo iptables -I OUTPUT -j FABLAB_OUTPUT)"
}

func blockIncomingCmd(port uint16) string {
	return fmt.Sprintf("sudo iptables -A FABLAB_INPUT -p tcp --dport %d -j DROP", port)
}

func unblockIncomingCmd(port uint16) string {
	return fmt.Sprintf("sudo iptables -D FABLAB_INPUT -p tcp --dport %d -j DROP", port)
}

func blockOutgoingCmd(ip string, port uint16) string {
	return fmt.Sprintf("sudo iptables -A FABLAB_OUTPUT -p tcp -d %s --dport %d -j DROP", ip, port)
}

func unblockOutgoingCmd(ip string, port uint16) string {
	return fmt.Sprintf("sudo iptables -D FABLAB_OUTPUT -p tcp -d %s --dport %d -j DROP", ip, port)
}

func unblockAllCmd() string {
	return "sudo iptables -F FABLAB_INPUT 2>/dev/null || true" +
		" && sudo iptables -F FABLAB_OUTPUT 2>/dev/null || true"
}

func killIncomingCmd(port uint16) string {
	return fmt.Sprintf("sudo ss -K -t state established sport = :%d", port)
}

func killOutgoingCmd(ip string, port uint16) string {
	return fmt.Sprintf("sudo ss -K -t state established dst %s:%d", ip, port)
}

// ensureFablabChains idempotently creates the FABLAB_INPUT and FABLAB_OUTPUT
// iptables chains and inserts jump rules from the built-in INPUT/OUTPUT chains.
func (host *Host) ensureFablabChains() error {
	if output, err := host.ExecLogged(ensureFablabChainsCmd()); err != nil {
		return fmt.Errorf("failed to ensure fablab iptables chains on [%s]: %w (output: %s)", host.PublicIp, err, output)
	}
	return nil
}

// BlockIncoming adds an iptables rule to DROP all incoming TCP traffic on the given port.
func (host *Host) BlockIncoming(port uint16) error {
	if err := host.ensureFablabChains(); err != nil {
		return err
	}

	if output, err := host.ExecLogged(blockIncomingCmd(port)); err != nil {
		return fmt.Errorf("failed to block incoming port %d on [%s]: %w (output: %s)", port, host.PublicIp, err, output)
	}
	return nil
}

// UnblockIncoming removes the iptables rule that DROPs incoming TCP traffic on the given port.
// This is best-effort: if the rule doesn't exist, a warning is logged but no error is returned.
func (host *Host) UnblockIncoming(port uint16) error {
	if output, err := host.ExecLogged(unblockIncomingCmd(port)); err != nil {
		logrus.WithField("hostId", host.Id).Warnf("failed to unblock incoming port %d on [%s]: %v (output: %s)", port, host.PublicIp, err, output)
	}
	return nil
}

// BlockOutgoing adds an iptables rule to DROP all outgoing TCP traffic to the given ip and port.
func (host *Host) BlockOutgoing(ip string, port uint16) error {
	if err := host.ensureFablabChains(); err != nil {
		return err
	}

	if output, err := host.ExecLogged(blockOutgoingCmd(ip, port)); err != nil {
		return fmt.Errorf("failed to block outgoing %s:%d on [%s]: %w (output: %s)", ip, port, host.PublicIp, err, output)
	}
	return nil
}

// UnblockOutgoing removes the iptables rule that DROPs outgoing TCP traffic to the given ip and port.
// This is best-effort: if the rule doesn't exist, a warning is logged but no error is returned.
func (host *Host) UnblockOutgoing(ip string, port uint16) error {
	if output, err := host.ExecLogged(unblockOutgoingCmd(ip, port)); err != nil {
		logrus.WithField("hostId", host.Id).Warnf("failed to unblock outgoing %s:%d on [%s]: %v (output: %s)", ip, port, host.PublicIp, err, output)
	}
	return nil
}

// UnblockAll flushes all rules from the FABLAB_INPUT and FABLAB_OUTPUT chains.
// This is best-effort: if the chains don't exist, a warning is logged but no error is returned.
func (host *Host) UnblockAll() error {
	if output, err := host.ExecLogged(unblockAllCmd()); err != nil {
		logrus.WithField("hostId", host.Id).Warnf("failed to unblock all on [%s]: %v (output: %s)", host.PublicIp, err, output)
	}
	return nil
}

// KillIncoming kills established TCP connections on the given local port using ss -K.
func (host *Host) KillIncoming(port uint16) error {
	if output, err := host.ExecLogged(killIncomingCmd(port)); err != nil {
		return fmt.Errorf("failed to kill incoming connections on port %d on [%s]: %w (output: %s)", port, host.PublicIp, err, output)
	}
	return nil
}

// KillOutgoing kills established TCP connections to the given remote ip and port using ss -K.
func (host *Host) KillOutgoing(ip string, port uint16) error {
	if output, err := host.ExecLogged(killOutgoingCmd(ip, port)); err != nil {
		return fmt.Errorf("failed to kill outgoing connections to %s:%d on [%s]: %w (output: %s)", ip, port, host.PublicIp, err, output)
	}
	return nil
}

// DisruptIncoming kills existing TCP connections on the given port and blocks new ones.
func (host *Host) DisruptIncoming(port uint16) error {
	if err := host.KillIncoming(port); err != nil {
		return err
	}
	return host.BlockIncoming(port)
}

// DisruptOutgoing kills existing TCP connections to the given ip:port and blocks new ones.
func (host *Host) DisruptOutgoing(ip string, port uint16) error {
	if err := host.KillOutgoing(ip, port); err != nil {
		return err
	}
	return host.BlockOutgoing(ip, port)
}
