package parallel

import (
	"fmt"
	"testing"
)

func Test_semaphore(t *testing.T) {
	var tasks []Task
	tasks = append(tasks, func() error {
		fmt.Println("hello")
		return nil
	})

	tasks = append(tasks, func() error {
		fmt.Println("world")
		return nil
	})

	Execute(tasks, 2)
}
