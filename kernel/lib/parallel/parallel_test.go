package parallel

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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

	err := Execute(tasks, 2)
	assert.NoError(t, err)
}
