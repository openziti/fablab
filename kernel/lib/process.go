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

package lib

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"sync"
)

func NewProcess(name string, cmd ...string) *Process {
	return &Process{
		Cmd:       exec.Command(name, cmd...),
		outStream: make(chan []byte),
		errStream: make(chan []byte),
	}
}

func (prc *Process) WithTail(tail TailFunction) *Process {
	prc.tail = tail
	return prc
}

func (prc *Process) Run() error {
	stdout, err := prc.Cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error getting stdout pipe (%s)", err)
	}

	stderr, err := prc.Cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error getting stderr pipe (%s)", err)
	}

	logrus.Infof("executing %v", prc.Cmd.Args)

	if err := prc.Cmd.Start(); err != nil {
		return fmt.Errorf("error starting cmd (%s)", err)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	go prc.reader(stdout, prc.outStream, &wg)
	go prc.combiner(&wg)
	prc.reader(stderr, prc.errStream, &wg)

	wg.Wait()

	if err := prc.Cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting (%s)", err)
	}

	return nil
}

type Process struct {
	Cmd       *exec.Cmd
	Output    bytes.Buffer
	outStream chan []byte
	errStream chan []byte
	tail      TailFunction
}

func StdoutTail(data []byte) {
	fmt.Printf(string(data))
}

type TailFunction func(data []byte)

func (prc *Process) combiner(wg *sync.WaitGroup) {
	defer wg.Done()

	outDone := false
	errDone := false
	for {
		select {
		case data := <-prc.outStream:
			if data != nil {
				prc.Output.Write(data)
				if prc.tail != nil {
					prc.tail(data)
				}
			} else {
				outDone = true
			}
		case data := <-prc.errStream:
			if data != nil {
				prc.Output.Write(data)
				if prc.tail != nil {
					prc.tail(data)
				}
			} else {
				errDone = true
			}
		}
		if outDone && errDone {
			return
		}
	}
}

func (prc *Process) reader(r io.ReadCloser, o chan []byte, wg *sync.WaitGroup) {
	defer close(o)
	defer wg.Done()

	for {
		buf := make([]byte, 64*1024)
		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Printf("error reading (%s)\n", err)
			return
		}
		if n > 0 {
			o <- buf[:n]
		}
	}
}
