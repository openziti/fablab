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

package subcmd

import (
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/lib/figlet"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
	"time"
)

func init() {
	RootCmd.AddCommand(newExecLoopCmd())
}

type execLoopCmd struct {
	bindings []string
}

func newExecLoopCmd() *cobra.Command {
	execLoop := &execLoopCmd{}

	cobraCmd := &cobra.Command{
		Use:   "exec-loop <until> <action> [<actions>...]",
		Short: "execute one or more actions",
		Example: "fablab exec-loop forever make-changes validate\n" +
			"fablab exec-loop 100 make-changes validate\n" +
			"fablab exec-loop 10m make-changes validate",
		Args: cobra.MinimumNArgs(2),
		Run:  execLoop.runExec,
	}

	cobraCmd.Flags().StringArrayVarP(&execCmdBindings, "variable", "b", []string{}, "specify variable binding ('<hostSpec>.a.b.c=value')")

	return cobraCmd
}

func (self *execLoopCmd) runExec(_ *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	ctx, err := model.NewRun()
	if err != nil {
		logrus.WithError(err).Fatal("error initializing run")
	}

	m := model.GetModel()

	if !m.IsBound() {
		logrus.Fatalf("model not bound")
	}

	for _, binding := range self.bindings {
		if err := execCmdBind(m, binding); err != nil {
			logrus.Fatalf("error binding [%s] (%v)", binding, err)
		}
	}

	var actions []model.Action

	for _, name := range args[1:] {
		action, found := m.GetAction(name)
		if !found {
			logrus.Fatalf("no such action [%s]", name)
		}
		actions = append(actions, action)
	}

	until, err := self.parseUntil(args[0])
	if err != nil {
		logrus.Fatalf("invalid until specification, must 'forever', a number (iterations) or a duration [%s]", args[0])
	}

	iterations := 1
	start := time.Now()

	for {
		iterationStart := time.Now()
		figlet.Figlet(fmt.Sprintf("ITERATION-%03d", iterations))
		for _, action := range actions {
			if err = action.Execute(ctx); err != nil {
				logrus.WithError(err).Fatalf("action failed [%+v]", action)
			}
		}
		if until.isDone() {
			pfxlog.Logger().Infof("finished after %v iteration(s) in %v", iterations, time.Since(start))
			return
		}
		pfxlog.Logger().Infof("iteration: %v, iteration time: %v, total time: %v",
			iterations, time.Since(iterationStart), time.Since(start))
		iterations++
	}
}

func (self *execLoopCmd) parseUntil(v string) (untilPredicate, error) {
	if strings.EqualFold(v, "forever") {
		return untilForever{}, nil
	}
	if v, err := strconv.Atoi(v); err == nil {
		return &untilIterations{
			limit: v,
		}, nil
	}
	if d, err := time.ParseDuration(v); err == nil {
		return &untilDeadline{
			deadline: time.Now().Add(d),
		}, nil
	}
	return nil, fmt.Errorf("invalid until spec '%s'", v)
}

type untilPredicate interface {
	isDone() bool
}

type untilIterations struct {
	limit int
	count int
}

func (self *untilIterations) isDone() bool {
	self.count++
	return self.count >= self.limit
}

type untilDeadline struct {
	deadline time.Time
}

func (self *untilDeadline) isDone() bool {
	return time.Now().After(self.deadline)
}

type untilForever struct{}

func (self untilForever) isDone() bool {
	return false
}
