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

package subcmd

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseScpArg(t *testing.T) {
	cases := []struct {
		name string
		arg  string
		want scpArg
	}{
		{"absolute local", "/var/log/x", scpArg{path: "/var/log/x"}},
		{"dot-relative local", "./logs/x", scpArg{path: "./logs/x"}},
		{"dot-dot-relative local", "../x", scpArg{path: "../x"}},
		{"bare local", "file.txt", scpArg{path: "file.txt"}},
		{"remote with path", "ctrl1:./logs/x", scpArg{hostSpec: "ctrl1", path: "./logs/x"}},
		{"remote absolute path", "ctrl1:/var/log/x", scpArg{hostSpec: "ctrl1", path: "/var/log/x"}},
		{"remote home shorthand", "ctrl1:", scpArg{hostSpec: "ctrl1", path: "."}},
		{"only first colon splits", "host:a:b", scpArg{hostSpec: "host", path: "a:b"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.want, parseScpArg(c.arg))
		})
	}
}

func TestParseScpArgWindowsVolumePaths(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows volume-path handling only applies on windows builds")
	}
	// Every volume-qualified form must be treated as local, not as a remote selector.
	for _, arg := range []string{`C:\work\file`, `C:/work/file`, `C:relative`, `\\?\C:\long`, `\\server\share\file`} {
		t.Run(arg, func(t *testing.T) {
			got := parseScpArg(arg)
			require.False(t, got.isRemote(), "expected local")
			require.Equal(t, arg, got.path)
		})
	}
}
