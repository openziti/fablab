/*
	Copyright 2020 NetFoundry, Inc.

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
	"archive/zip"
	"fmt"
	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/foundation/util/info"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

func Export(path string, m *model.Model) error {
	zipFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating archive [%s] (%w)", path, err)
	}
	defer func() { _ = zipFile.Close() }()

	zip := zip.NewWriter(zipFile)
	defer func() { _ = zip.Close() }()

	root := model.ActiveInstancePath()

	paths := []string{
		filepath.Join(root, "dumps"),
		filepath.Join(root, "reports"),
		filepath.Join(root, "forensics"),
	}
	for _, path := range paths {
		if fi, err := os.Stat(path); err == nil {
			if fi.IsDir() {
				scanner := newInstanceScanner(root, zip)
				if err := filepath.Walk(path, scanner.visit); err != nil {
					return fmt.Errorf("error walking [%s] (%w)", path, err)
				}
			}
		}
	}

	return nil
}

func newInstanceScanner(root string, zip *zip.Writer) *instanceScanner {
	return &instanceScanner{
		root: root,
		zip:  zip,
	}
}

func (self *instanceScanner) visit(path string, fi os.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error walking (%w)", err)
	}

	if !fi.IsDir() {
		file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error opening [%s] (%w)", path, err)
		}
		defer func() { _ = file.Close() }()

		fi, err := file.Stat()
		if err != nil {
			return fmt.Errorf("error stat-ing file [%s] (%w)", path, err)
		}

		header, err := zip.FileInfoHeader(fi)
		if err != nil {
			return fmt.Errorf("error creating fileinfo header [%s] (%w)", path, err)
		}
		basepath, err := filepath.Rel(self.root, path)
		if err != nil {
			return fmt.Errorf("error computing basepath [%s] (%w)", path, err)
		}
		header.Name = basepath
		header.Method = zip.Deflate

		writer, err := self.zip.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("error creating zip header [%s] (%w)", path, err)
		}

		n, err := io.Copy(writer, file)
		if err != nil {
			return fmt.Errorf("error copying file [%s] (%w)", path, err)
		}
		logrus.Infof("=> %s (%s)", path, info.ByteCount(n))
	}

	return nil
}

type instanceScanner struct {
	root string
	zip  *zip.Writer
}
