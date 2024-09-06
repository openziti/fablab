package parallel

import (
	"context"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/lib/util"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
	"sync/atomic"
)

type Task func() error

func Execute(tasks []Task, concurrency int64) error {
	if len(tasks) == 0 {
		pfxlog.Logger().Warn("ran parallel set of tasks, but no tasks provided")
		return nil
	}

	if concurrency < 1 {
		return errors.Errorf("invalid concurrency %v, must be at least 1", concurrency)
	}

	completed := atomic.Int64{}

	sem := semaphore.NewWeighted(concurrency)
	errorsC := make(chan error, len(tasks))
	for _, task := range tasks {
		if err := sem.Acquire(context.Background(), 1); err != nil {
			errorsC <- err
			continue
		}
		boundTask := task
		go func() {
			defer func() {
				sem.Release(1)
				current := completed.Add(1)
				if current%10 == 0 {
					pfxlog.Logger().Infof("completed %d/%d tasks", current, len(tasks))
				}
				if int(current) == len(tasks) {
					close(errorsC)
				}
			}()
			if err := boundTask(); err != nil {
				errorsC <- err
			}
		}()
	}

	var errList []error
	for err := range errorsC {
		errList = append(errList, err)
	}

	if len(errList) == 0 {
		return nil
	}

	if len(errList) == 1 {
		return errList[0]
	}

	return util.MultipleErrors(errList)
}
