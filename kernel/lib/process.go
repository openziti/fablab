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
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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

// WithOutputFile has Run send the command's stdout and stderr to the file at path, created or
// truncated, rather than to pipes, and tail that file into Output and the tail function. Because
// the command doesn't write to anything this process reads, it keeps running and writing to the file
// if this process exits first.
func (prc *Process) WithOutputFile(path string) *Process {
	prc.outputPath = path
	return prc
}

func (prc *Process) Run() error {
	if prc.outputPath != "" {
		return prc.runWithOutputFile()
	}

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

// outputPollInterval is how long runWithOutputFile waits for the command to write more once it has
// read everything written so far.
const outputPollInterval = 100 * time.Millisecond

// runWithOutputFile runs the command with its stdout and stderr on the file at outputPath, reading
// the file as it grows until the command exits.
func (prc *Process) runWithOutputFile() error {
	out, err := os.Create(prc.outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file (%w)", err)
	}
	in, err := os.Open(prc.outputPath)
	if err != nil {
		_ = out.Close()
		return fmt.Errorf("error opening output file (%w)", err)
	}
	defer func() { _ = in.Close() }()

	prc.Cmd.Stdout = out
	prc.Cmd.Stderr = out

	logrus.Infof("executing %v, output in %s", prc.Cmd.Args, prc.outputPath)

	err = prc.Cmd.Start()
	_ = out.Close()
	if err != nil {
		return fmt.Errorf("error starting cmd (%s)", err)
	}

	exitedC := make(chan error, 1)
	go func() {
		exitedC <- prc.Cmd.Wait()
	}()

	waitErr, readErr := prc.tailUntilExit(in, exitedC)
	var errs []error
	if waitErr != nil {
		errs = append(errs, fmt.Errorf("error waiting (%s)", waitErr))
	}
	if readErr != nil {
		errs = append(errs, fmt.Errorf("error reading output file %s, Output is incomplete (%w)", prc.outputPath, readErr))
	}
	return errors.Join(errs...)
}

// tailUntilExit emits what r holds as it grows, until exitedC delivers the command's exit and r has
// nothing more. It returns the command's exit error, plus the read error if reading r failed first;
// either way it waits for the exit.
func (prc *Process) tailUntilExit(r io.Reader, exitedC <-chan error) (waitErr, readErr error) {
	exited := false
	buf := make([]byte, 64*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			prc.emit(buf[:n])
		}
		switch {
		case err == nil:
		case err != io.EOF:
			if !exited {
				waitErr = <-exitedC
			}
			return waitErr, err
		case exited:
			return waitErr, nil
		default:
			select {
			case waitErr = <-exitedC:
				exited = true
			case <-time.After(outputPollInterval):
			}
		}
	}
}

type Process struct {
	Cmd        *exec.Cmd
	Output     bytes.Buffer
	outStream  chan []byte
	errStream  chan []byte
	tail       TailFunction
	outputPath string
}

func (prc *Process) emit(data []byte) {
	prc.Output.Write(data)
	if prc.tail != nil {
		prc.tail(data)
	}
}

func StdoutTail(data []byte) {
	fmt.Print(string(data))
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
				prc.emit(data)
			} else {
				outDone = true
			}
		case data := <-prc.errStream:
			if data != nil {
				prc.emit(data)
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
