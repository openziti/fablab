package fablib

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	"io"
	"log"
	"strings"
	"unicode"
)

func SummarizeSar(data []byte) (*model.HostSummary, error) {
	summary := &model.HostSummary{}

	reader := bufio.NewReader(bytes.NewBuffer(sarTestData[:]))
	var err error
	line := ""
	for line != "\n" {
		line, err = reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("premature eof")
			} else {
				return nil, fmt.Errorf("unknown error (%w)", err)
			}
		}
		log.Printf("read [%s]", line)
	}
	log.Printf("read header, reading data")

	for {
		headerLine, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				return nil, nil
			} else {
				return nil, fmt.Errorf("unexpected error (%w)", err)
			}
		}
		headerTokens := strings.FieldsFunc(headerLine[11:], func(c rune) bool {
			return unicode.IsSpace(c)
		})
		log.Printf("header tokens [%q]", headerTokens)

		dataLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("unexpected error (%w)", err)
		}
		dataTimeStr := dataLine[:11]
		log.Printf("data time [%s]", dataTimeStr)
		dataTokens := strings.FieldsFunc(dataLine[11:], func(c rune) bool {
			return unicode.IsSpace(c)
		})
		log.Printf("data tokens [%q]", dataTokens)

		blankLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("unexpected error (%w)", err)
		}
		if blankLine != "\n" {
			return nil, fmt.Errorf("synchronization error [%s]", blankLine)
		}
		log.Printf("blank")

		switch headerTokens[0] {
		case "CPU":
			cpu, err := cpuTimeslice(headerTokens, dataTokens)
			if err != nil {
				return nil, fmt.Errorf("error processing cpu timeslice (%w)", err)
			}
			summary.Cpu = append(summary.Cpu, cpu)

		case "kbmemfree":
			memory, err := memoryTimeslice(headerTokens, dataTokens)
			if err != nil {
				return nil, fmt.Errorf("error processing memory timeslice (%w)", err)
			}
			summary.Memory = append(summary.Memory, memory)

		case "runq-sz":
			process, err := processTimeslice(headerTokens, dataTokens)
			if err != nil {
				return nil, fmt.Errorf("error processing process timeslice (%w)", err)
			}
			summary.Process = append(summary.Process, process)

		default:
			return nil, fmt.Errorf("unrecognized header token[%s]", headerTokens[0])
		}
	}
}

func cpuTimeslice(header, data []string) (*model.CpuTimeslice, error) {
	return &model.CpuTimeslice{}, nil
}

func memoryTimeslice(header, data []string) (*model.MemoryTimeslice, error) {
	return &model.MemoryTimeslice{}, nil
}

func processTimeslice(header, data []string) (*model.ProcessTimeslice, error) {
	return &model.ProcessTimeslice{}, nil
}