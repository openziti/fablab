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

// Package sshtest provides an in-process SSH server exposing an SFTP subsystem,
// backed by a temporary directory, for exercising the libssh helpers in tests
// without a real fablab host.
package sshtest

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Factory is an libssh.SshConfigFactory implementation that points at a Server. It is
// returned as a concrete type so this package does not depend on libssh; it satisfies
// the interface structurally at the call site.
type Factory struct {
	host   string
	port   int
	user   string
	signer ssh.Signer
}

func (f *Factory) Address() string  { return net.JoinHostPort(f.host, strconv.Itoa(f.port)) }
func (f *Factory) Hostname() string { return f.host }
func (f *Factory) Port() int        { return f.port }
func (f *Factory) User() string     { return f.user }
func (f *Factory) KeyPath() string  { return "" }

func (f *Factory) Config() *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User:            f.user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(f.signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}

// Server is an in-process SSH+SFTP server listening on the loopback interface. Its
// SFTP subsystem operates on the real filesystem; tests use paths under Root.
type Server struct {
	// Root is a temporary directory that stands in for the remote host's filesystem.
	Root string

	listener net.Listener
	factory  *Factory
}

// Start launches a server on 127.0.0.1 with a freshly generated host and client
// keypair, serving SFTP rooted (by convention) at a temporary directory. It registers
// cleanup with t and returns the ready server.
func Start(t *testing.T) *Server {
	t.Helper()

	clientSigner := mustSigner(t)
	hostSigner := mustSigner(t)
	clientKey := clientSigner.PublicKey().Marshal()

	serverConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(_ ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if bytes.Equal(key.Marshal(), clientKey) {
				return &ssh.Permissions{}, nil
			}
			return nil, fmt.Errorf("unknown public key")
		},
	}
	serverConfig.AddHostKey(hostSigner)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to listen: %v", err)
	}

	host, portStr, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("unable to parse listener address: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("unable to parse listener port: %v", err)
	}

	s := &Server{
		Root:     t.TempDir(),
		listener: listener,
		factory:  &Factory{host: host, port: port, user: "test", signer: clientSigner},
	}

	go s.acceptLoop(serverConfig)
	t.Cleanup(func() { _ = listener.Close() })

	return s
}

// Factory returns a config factory that connects to this server.
func (s *Server) Factory() *Factory { return s.factory }

func (s *Server) acceptLoop(cfg *ssh.ServerConfig) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return // listener closed
		}
		go handleConn(conn, cfg)
	}
}

func handleConn(nConn net.Conn, cfg *ssh.ServerConfig) {
	defer func() { _ = nConn.Close() }()

	sshConn, chans, reqs, err := ssh.NewServerConn(nConn, cfg)
	if err != nil {
		return
	}
	defer func() { _ = sshConn.Close() }()
	go ssh.DiscardRequests(reqs)

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			_ = newChan.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChan.Accept()
		if err != nil {
			continue
		}
		go handleSession(channel, requests)
	}
}

func handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	for req := range requests {
		// The only request needed to exercise the libssh helpers is the sftp subsystem.
		if req.Type == "subsystem" && len(req.Payload) >= 4 && string(req.Payload[4:]) == "sftp" {
			_ = req.Reply(true, nil)
			server, err := sftp.NewServer(channel)
			if err != nil {
				_ = channel.Close()
				return
			}
			_ = server.Serve()
			_ = server.Close()
			_ = channel.Close()
			return
		}
		_ = req.Reply(false, nil)
	}
}

func mustSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("unable to generate key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("unable to create signer: %v", err)
	}
	return signer
}
