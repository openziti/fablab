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

package libssh_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openziti/fablab/internal/sshtest"
	"github.com/openziti/fablab/kernel/libssh"
	"github.com/stretchr/testify/require"
)

// makeRemovable ensures every directory under root is writable at cleanup time, so a
// read-only directory a test creates does not break t.TempDir's RemoveAll. It must be
// registered after the t.TempDir it covers so it runs first (cleanups are LIFO).
func makeRemovable(t *testing.T, root string) {
	t.Cleanup(func() {
		_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err == nil && d.IsDir() {
				_ = os.Chmod(p, 0o755)
			}
			return nil
		})
	})
}

// writeFile writes content to path with an exact mode, independent of the process
// umask, so mode-preservation assertions are deterministic.
func writeFile(t *testing.T, path string, content string, mode os.FileMode) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), mode))
	require.NoError(t, os.Chmod(path, mode))
}

func TestRetrieveRemoteFilesCopiesDirectoryContents(t *testing.T) {
	srv := sshtest.Start(t)

	remoteDir := filepath.Join(srv.Root, "logs")
	writeFile(t, filepath.Join(remoteDir, "one.log"), "1", 0o644)
	writeFile(t, filepath.Join(remoteDir, "nested", "two.log"), "2", 0o640)

	localTarget := filepath.Join(t.TempDir(), "dst")
	require.NoError(t, libssh.RetrieveRemoteFiles(srv.Factory(), localTarget, remoteDir))

	// RetrieveRemoteFiles copies the directory's contents into localTarget.
	one, err := os.ReadFile(filepath.Join(localTarget, "one.log"))
	require.NoError(t, err)
	require.Equal(t, "1", string(one))

	fi, err := os.Stat(filepath.Join(localTarget, "one.log"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o644), fi.Mode().Perm())

	two, err := os.ReadFile(filepath.Join(localTarget, "nested", "two.log"))
	require.NoError(t, err)
	require.Equal(t, "2", string(two))
}

func TestRetrieveRemoteFilesPreservesNestedDirModes(t *testing.T) {
	srv := sshtest.Start(t)

	remoteDir := filepath.Join(srv.Root, "src")
	require.NoError(t, os.MkdirAll(filepath.Join(remoteDir, "sub"), 0o755))
	writeFile(t, filepath.Join(remoteDir, "sub", "f"), "x", 0o644)
	require.NoError(t, os.Chmod(filepath.Join(remoteDir, "sub"), 0o750))
	require.NoError(t, os.Chmod(remoteDir, 0o711))

	localTarget := filepath.Join(t.TempDir(), "dst")
	require.NoError(t, libssh.RetrieveRemoteFiles(srv.Factory(), localTarget, remoteDir))

	// The destination container is not a mirror of the source dir, so it must not take
	// the source's 0711; nested subdirectories are mirrors and keep their modes.
	top, err := os.Stat(localTarget)
	require.NoError(t, err)
	require.NotEqual(t, os.FileMode(0o711), top.Mode().Perm(), "container must not take the source dir's mode")

	subInfo, err := os.Stat(filepath.Join(localTarget, "sub"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o750), subInfo.Mode().Perm())
}

func TestRetrieveRemoteFilesMultiSourceReadOnlyDirThenFile(t *testing.T) {
	srv := sshtest.Start(t)
	makeRemovable(t, srv.Root)

	// A read-only source directory, plus a separate file.
	roDir := filepath.Join(srv.Root, "rodir")
	writeFile(t, filepath.Join(roDir, "inner"), "i", 0o644)
	require.NoError(t, os.Chmod(roDir, 0o555))
	other := filepath.Join(srv.Root, "other.txt")
	writeFile(t, other, "o", 0o644)

	dst := filepath.Join(t.TempDir(), "dst")
	// Dir first, then file: the container must stay writable so the file can be added,
	// i.e. the container must not have inherited the read-only dir's 0555.
	require.NoError(t, libssh.RetrieveRemoteFiles(srv.Factory(), dst, roDir, other))

	inner, err := os.ReadFile(filepath.Join(dst, "inner"))
	require.NoError(t, err)
	require.Equal(t, "i", string(inner))
	o, err := os.ReadFile(filepath.Join(dst, "other.txt"))
	require.NoError(t, err)
	require.Equal(t, "o", string(o))
}

func TestRetrieveRemoteFilesLeavesExistingDirMode(t *testing.T) {
	srv := sshtest.Start(t)
	remoteDir := filepath.Join(srv.Root, "src")
	writeFile(t, filepath.Join(remoteDir, "f"), "x", 0o644)
	require.NoError(t, os.Chmod(remoteDir, 0o700))

	dst := t.TempDir()                       // pre-existing destination
	require.NoError(t, os.Chmod(dst, 0o755)) // distinct from the remote dir's 0700

	require.NoError(t, libssh.RetrieveRemoteFiles(srv.Factory(), dst, remoteDir))

	fi, err := os.Stat(dst)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o755), fi.Mode().Perm(), "pre-existing destination dir mode must be preserved")
	got, err := os.ReadFile(filepath.Join(dst, "f"))
	require.NoError(t, err)
	require.Equal(t, "x", string(got))
}

func TestRetrieveRemoteFilesReadOnlySource(t *testing.T) {
	srv := sshtest.Start(t)
	makeRemovable(t, srv.Root)
	remoteDir := filepath.Join(srv.Root, "src")
	ro := filepath.Join(remoteDir, "ro")
	writeFile(t, filepath.Join(ro, "f"), "x", 0o644)
	require.NoError(t, os.Chmod(ro, 0o555)) // read-only: children must be created before chmod

	base := t.TempDir()
	makeRemovable(t, base)
	localTarget := filepath.Join(base, "dst")
	require.NoError(t, libssh.RetrieveRemoteFiles(srv.Factory(), localTarget, remoteDir))

	fi, err := os.Stat(filepath.Join(localTarget, "ro"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o555), fi.Mode().Perm())
	got, err := os.ReadFile(filepath.Join(localTarget, "ro", "f"))
	require.NoError(t, err)
	require.Equal(t, "x", string(got))
}

func TestRetrieveRemoteFilesFollowsDirSymlink(t *testing.T) {
	srv := sshtest.Start(t)

	remoteDir := filepath.Join(srv.Root, "src")
	real := filepath.Join(remoteDir, "real")
	writeFile(t, filepath.Join(real, "inside.txt"), "hello", 0o644)
	require.NoError(t, os.Symlink(real, filepath.Join(remoteDir, "link")))

	localTarget := filepath.Join(t.TempDir(), "dst")
	require.NoError(t, libssh.RetrieveRemoteFiles(srv.Factory(), localTarget, remoteDir))

	got, err := os.ReadFile(filepath.Join(localTarget, "link", "inside.txt"))
	require.NoError(t, err)
	require.Equal(t, "hello", string(got))
}
