package fablib

import "github.com/openziti/fabric/controller/network"

type Task func() error

func InParallel(tasks ...Task) error {
	errorsC := make(chan error, len(tasks))
	var joiners []chan struct{}
	for _, task := range tasks {
		boundTask := task
		joiner := make(chan struct{})
		joiners = append(joiners, joiner)
		go func() {
			defer close(joiner)
			if err := boundTask(); err != nil {
				errorsC <- err
			}
		}()
	}

	for _, joiner := range joiners {
		<-joiner
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

	return network.MultipleErrors(errors)
}
