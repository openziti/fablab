/*
	Copyright 2019 Netfoundry, Inc.

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

package internal

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

func CopyFile(srcPath, dstPath string) (int64, error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = srcFile.Close() }()

	srcStat, err := srcFile.Stat()
	if err != nil {
		return 0, err
	}

	if !srcStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", srcPath)
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = dstFile.Close() }()

	logrus.Debugf("[%s] => [%s]", srcPath, dstPath)

	return io.Copy(dstFile, srcFile)
}

func CopyTree(srcPath, dstPath string) error {
	visitor := &copyTreeVisitor{srcPath: srcPath, dstPath: dstPath}
	if err := filepath.Walk(srcPath, visitor.visit); err != nil {
		return fmt.Errorf("error scanning source tree [%s] (%w)", srcPath, err)
	}
	return nil
}

func (copyTreeVisitor *copyTreeVisitor) visit(path string, fi os.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error in visitor (%w)", err)
	}

	if fi.Mode().IsRegular() {
		rel, err := filepath.Rel(copyTreeVisitor.srcPath, path)
		if err != nil {
			return fmt.Errorf("error relativizing path [%s] (%w)", copyTreeVisitor.srcPath, err)
		}

		dstPath := filepath.Join(copyTreeVisitor.dstPath, rel)

		if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
			return fmt.Errorf("error creating parent directories for [%s] (%w)", dstPath, err)
		}

		if _, err := CopyFile(path, dstPath); err != nil {
			return fmt.Errorf("error copying [%s] => [%s] (%w)", path, dstPath, err)
		}

		logrus.Infof("[%s] => [%s]", path, dstPath)
	}

	return nil
}

type copyTreeVisitor struct {
	srcPath string
	dstPath string
}
