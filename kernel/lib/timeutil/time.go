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

package timeutil

import "time"

func TimeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / 1000000
}

func MillisecondsToTime(millis int64) time.Time {
	return time.Unix(0, millis*time.Millisecond.Nanoseconds())
}
