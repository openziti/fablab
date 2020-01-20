/*
	Copyright 2019 NetFoundry, Inc.

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

package terraform_0

// terraformForwardSlashPath returns paths that use forward slashes only, even on Windows. Terraform
// files that contain partial paths should use the forward slash such that they concatenate with
// forward slashes only. On Windows this means that Terraform will expect paths such as C:/my/path
// and will process them correctly.
func terraformForwardSlashPath(path string) {
	return path //no-op on linux systems using forward slashes
}
