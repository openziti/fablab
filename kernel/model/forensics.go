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
	"os"
	"path/filepath"
	"strings"
)

func AllocateForensicScenario(run, scenario string) string {
	return fmt.Sprintf("%s/forensics/%s/%s", BuildPath(), run, scenario)
}

func AllocateDump(run string) string {
	return fmt.Sprintf("%s/dumps/%s.json", BuildPath(), run)
}

func ListDumps() ([]string, error) {
	files, err := os.ReadDir(filepath.Join(BuildPath(), "dumps"))
	if err != nil {
		return nil, fmt.Errorf("unable to list dumps (%w)", err)
	}

	var dumps []string
	for _, file := range files {
		if file.Type().IsRegular() && strings.HasSuffix(file.Name(), ".json") {
			dumps = append(dumps, filepath.Join(BuildPath(), "dumps", file.Name()))
		}
	}

	return dumps, nil
}
