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

package fablib

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func SummarizeSar(data []byte) (*model.HostSummary, error) {
	summary := &model.HostSummary{}

	reader := bufio.NewReader(bytes.NewBuffer(data[:]))
	title, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("unexpected error (%w)", err)
	}
	titleTokens := strings.FieldsFunc(title, func(c rune) bool {
		return unicode.IsSpace(c)
	})
	var dateStr string
	if len(titleTokens) < 4 {
		return nil, fmt.Errorf("unexpected sar title format [%d]", len(titleTokens))
	}
	dateStr = titleTokens[3]
	logrus.Debugf("date [%s]", dateStr)

	blankLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("unexpected error (%w)", err)
	}
	if blankLine != "\n" {
		return nil, fmt.Errorf("expected blank line")
	}

	for {
		headerLine, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				return summary, nil
			} else {
				return nil, fmt.Errorf("unexpected error (%w)", err)
			}
		}
		headerTokens := strings.FieldsFunc(headerLine[11:], func(c rune) bool {
			return unicode.IsSpace(c)
		})
		logrus.Debugf("header tokens [%q]", headerTokens)

		dataLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("unexpected error (%w)", err)
		}
		dataTimestampStr := dataLine[:11]
		logrus.Debugf("data time [%s]", dataTimestampStr)
		dataTokens := strings.FieldsFunc(dataLine[11:], func(c rune) bool {
			return unicode.IsSpace(c)
		})
		logrus.Debugf("data tokens [%q]", dataTokens)

		blankLine, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				return summary, nil
			} else {
				return nil, fmt.Errorf("unexpected error (%w)", err)
			}
		}
		if blankLine != "\n" {
			return nil, fmt.Errorf("synchronization error [%s]", blankLine)
		}
		logrus.Debugf("blank")

		switch headerTokens[0] {
		case "CPU":
			cpu, err := cpuTimeslice(dateStr, dataTimestampStr, headerTokens, dataTokens)
			if err != nil {
				return nil, fmt.Errorf("error processing cpu timeslice (%w)", err)
			}
			if cpu != nil {
				summary.Cpu = append(summary.Cpu, cpu)
			}

		case "kbmemfree":
			memory, err := memoryTimeslice(dateStr, dataTimestampStr, headerTokens, dataTokens)
			if err != nil {
				return nil, fmt.Errorf("error processing memory timeslice (%w)", err)
			}
			if memory != nil {
				summary.Memory = append(summary.Memory, memory)
			}

		case "runq-sz":
			process, err := processTimeslice(dateStr, dataTimestampStr, headerTokens, dataTokens)
			if err != nil {
				return nil, fmt.Errorf("error processing process timeslice (%w)", err)
			}
			if process != nil {
				summary.Process = append(summary.Process, process)
			}

		default:
			return nil, fmt.Errorf("unrecognized header token[%s]", headerTokens[0])
		}
	}
}

func cpuTimeslice(dateStr, timestampStr string, header, data []string) (*model.CpuTimeslice, error) {
	if !strings.HasPrefix(timestampStr, "Average:") {
		summary := &model.CpuTimeslice{}

		timestampMs, err := dateTimeStrToMs(dateStr, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing time (%w)", err)
		}
		summary.TimestampMs = timestampMs

		for i, h := range header {
			switch h {
			case "CPU":
				//

			case "%user":
				percentUser, err := strconv.ParseFloat(data[i], 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing %%user [%s] (%w)", data[i], err)
				}
				summary.PercentUser = percentUser

			case "%nice":
				percentNice, err := strconv.ParseFloat(data[i], 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing %%nice [%s] (%w)", data[i], err)
				}
				summary.PercentNice = percentNice

			case "%system":
				percentSystem, err := strconv.ParseFloat(data[i], 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing %%system [%s] (%w)", data[i], err)
				}
				summary.PercentSystem = percentSystem

			case "%iowait":
				percentIowait, err := strconv.ParseFloat(data[i], 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing %%iowait [%s] (%w)", data[i], err)
				}
				summary.PercentIowait = percentIowait

			case "%steal":
				percentSteal, err := strconv.ParseFloat(data[i], 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing %%steal [%s] (%w)", data[i], err)
				}
				summary.PercentSteal = percentSteal

			case "%idle":
				percentIdle, err := strconv.ParseFloat(data[i], 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing %%idle [%s] (%w)", data[i], err)
				}
				summary.PercentIdle = percentIdle

			default:
				return nil, fmt.Errorf("unknown cpu header token [%s]", h)
			}
		}

		return summary, nil

	} else {
		return nil, nil
	}
}

func memoryTimeslice(dateStr, timestampStr string, header, data []string) (*model.MemoryTimeslice, error) {
	if !strings.HasPrefix(timestampStr, "Average:") {
		summary := &model.MemoryTimeslice{}

		timestampMs, err := dateTimeStrToMs(dateStr, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing time (%w)", err)
		}
		summary.TimestampMs = timestampMs

		for i, h := range header {
			switch h {
			case "kbmemfree":
				kbmemfree, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing kbmemfree [%s] (%w)", data[i], err)
				}
				summary.MemFreeK = kbmemfree

			case "kbavail":
				kbavail, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing kbavail [%s] (%w)", data[i], err)
				}
				summary.AvailK = kbavail

			case "kbmemused":
				kbmemused, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing kbmemused [%s] (%w)", data[i], err)
				}
				summary.UsedK = kbmemused

			case "%memused":
				percentMemused, err := strconv.ParseFloat(data[i], 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing %%memused [%s] (%w)", data[i], err)
				}
				summary.UsedPercent = percentMemused

			case "kbbuffers":
				kbbuffers, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing kbbuffers [%s] (%w)", data[i], err)
				}
				summary.BuffersK = kbbuffers

			case "kbcached":
				kbcached, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing kbcached [%s] (%w)", data[i], err)
				}
				summary.CachedK = kbcached

			case "kbcommit":
				kbcommit, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing kbcommit [%s] (%w)", data[i], err)
				}
				summary.CommitK = kbcommit

			case "%commit":
				percentCommit, err := strconv.ParseFloat(data[i], 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing %%commit [%s] (%w)", data[i], err)
				}
				summary.CommitPercent = percentCommit

			case "kbactive":
				kbactive, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing kbactive [%s] (%w)", data[i], err)
				}
				summary.ActiveK = kbactive

			case "kbinact":
				kbinact, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing kbinact [%s] (%w)", data[i], err)
				}
				summary.InactiveK = kbinact

			case "kbdirty":
				kbdirty, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing kbdirty [%s] (%w)", data[i], err)
				}
				summary.DirtyK = kbdirty

			default:
				return nil, fmt.Errorf("unknown memory header token [%s]", h)
			}
		}

		return summary, nil

	} else {
		return nil, nil
	}
}

func processTimeslice(dateStr, timestampStr string, header, data []string) (*model.ProcessTimeslice, error) {
	if !strings.HasPrefix(timestampStr, "Average:") {
		summary := &model.ProcessTimeslice{}

		timestampMs, err := dateTimeStrToMs(dateStr, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing time (%w)", err)
		}
		summary.TimestampMs = timestampMs

		for i, h := range header {
			switch h {
			case "runq-sz":
				runqSz, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing runq-sz [%s] (%w)", data[i], err)
				}
				summary.RunQueueSize = runqSz

			case "plist-sz":
				plistSz, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing plist-sz [%s] (%w)", data[i], err)
				}
				summary.ProcessListSize = plistSz

			case "ldavg-1":
				ldavg1, err := strconv.ParseFloat(data[i], 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing ldavg-1 [%s] (%w)", data[i], err)
				}
				summary.LoadAverage1m = ldavg1

			case "ldavg-5":
				ldavg5, err := strconv.ParseFloat(data[i], 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing ldavg-5 [%s] (%w)", data[i], err)
				}
				summary.LoadAverage5m = ldavg5

			case "ldavg-15":
				ldavg15, err := strconv.ParseFloat(data[i], 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing ldavg-15 [%s] (%w)", data[i], err)
				}
				summary.LoadAverage15m = ldavg15

			case "blocked":
				blocked, err := strconv.ParseInt(data[i], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing blocked [%s] (%w)", data[i], err)
				}
				summary.Blocked = blocked

			default:
				return nil, fmt.Errorf("unknown process header token [%s]", h)
			}
		}

		return summary, nil

	} else {
		return nil, nil
	}
}

func dateTimeStrToMs(dateStr, timeStr string) (int64, error) {
	fullStr := fmt.Sprintf("%s %s", dateStr, timeStr)
	t, err := time.Parse("01/02/2006 03:04:05 PM", fullStr)
	if err != nil {
		return -1, fmt.Errorf("error parsing time [%s] (%w)", fullStr, err)
	}
	return t.UnixNano() / 1000000, nil
}
