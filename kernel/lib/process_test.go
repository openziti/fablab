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

package lib

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/stretchr/testify/require"
)

// outputFileHelperEnv names the output file when the test binary runs as the parent process in
// Test_ProcessOutputFileOutlivesParent.
const outputFileHelperEnv = "FABLAB_TEST_PROCESS_OUTPUT_FILE"

func skipWithoutShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("needs sh")
	}
}

// With an output file, Run collects stdout and stderr from the file into Output and the tail
// function, and reports the command's exit status.
func Test_ProcessOutputFileCollectsOutput(t *testing.T) {
	skipWithoutShell(t)
	req := require.New(t)
	path := filepath.Join(t.TempDir(), "out.log")

	var tailed strings.Builder
	prc := NewProcess("sh", "-c", "echo to stdout; sleep 0.3; echo to stderr >&2; exit 3")
	prc.WithTail(func(data []byte) { tailed.Write(data) }).WithOutputFile(path)

	err := prc.Run()
	req.ErrorContains(err, "exit status 3")

	const expected = "to stdout\nto stderr\n"
	req.Equal(expected, prc.Output.String())
	req.Equal(expected, tailed.String())
	written, err := os.ReadFile(path)
	req.NoError(err)
	req.Equal(expected, string(written))
}

// A failed read of the output file is returned with the command's exit, after waiting for it, so the
// caller knows Output is incomplete.
func Test_ProcessTailReturnsReadError(t *testing.T) {
	req := require.New(t)
	readErr := errors.New("read failed")
	exitedC := make(chan error, 1)
	exitedC <- nil

	prc := NewProcess("unused")
	waitErr, err := prc.tailUntilExit(io.MultiReader(strings.NewReader("partial"), iotest.ErrReader(readErr)), exitedC)
	req.NoError(waitErr)
	req.ErrorIs(err, readErr)
	req.Empty(exitedC, "returned without waiting for the command to exit")
	req.Equal("partial", prc.Output.String())
}

// A command run with an output file keeps running and writing to it after the process that ran it
// is killed.
func Test_ProcessOutputFileOutlivesParent(t *testing.T) {
	skipWithoutShell(t)
	if path := os.Getenv(outputFileHelperEnv); path != "" {
		_ = NewProcess("sh", "-c", "echo started; sleep 1; echo still writing").WithOutputFile(path).Run()
		return
	}

	req := require.New(t)
	path := filepath.Join(t.TempDir(), "out.log")

	parent := exec.Command(os.Args[0], "-test.run=^Test_ProcessOutputFileOutlivesParent$")
	parent.Env = append(os.Environ(), outputFileHelperEnv+"="+path)
	req.NoError(parent.Start())

	req.Eventually(func() bool {
		written, _ := os.ReadFile(path)
		return strings.Contains(string(written), "started")
	}, 5*time.Second, 10*time.Millisecond, "command never started")
	req.NoError(parent.Process.Kill())
	_ = parent.Wait()

	req.Eventually(func() bool {
		written, _ := os.ReadFile(path)
		return strings.Contains(string(written), "still writing")
	}, 5*time.Second, 50*time.Millisecond, "command stopped writing once its parent was killed")
}
