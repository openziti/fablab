package parallel

import (
	"context"
	"github.com/openziti/fablab/kernel/lib/util"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
)

type Task func() error

func Execute(tasks []Task, concurrency int64) error {
	if concurrency < 1 {
		return errors.Errorf("invalid concurrency %v, must be at least 1", concurrency)
	}

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
			}()
			if err := boundTask(); err != nil {
				errorsC <- err
			}
		}()
	}

	if err := sem.Acquire(context.Background(), concurrency); err != nil {
		return err
	}

	close(errorsC)

	var errors []error
	for err := range errorsC {
		errors = append(errors, err)
	}

	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return errors[0]
	}

	return util.MultipleErrors(errors)
}
