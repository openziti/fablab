package parallel

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func Test_DependsOn_WaitsForDependency(t *testing.T) {
	var order []string
	orderCh := make(chan string, 3)

	taskA := TaskWithLabel("test", "taskA", func() error {
		time.Sleep(50 * time.Millisecond)
		orderCh <- "A"
		return nil
	})

	taskB := TaskWithLabel("test", "taskB", func() error {
		orderCh <- "B"
		return nil
	})

	taskB.DependsOn(taskA, 5*time.Second)

	tasks := []LabeledTask{taskB, taskA}
	err := ExecuteLabeled(tasks, 2, AlwaysReport())
	require.NoError(t, err)

	close(orderCh)
	for v := range orderCh {
		order = append(order, v)
	}

	require.Equal(t, []string{"A", "B"}, order)
}

func Test_DependsOn_Timeout(t *testing.T) {
	taskA := TaskWithLabel("test", "taskA", func() error {
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	taskB := TaskWithLabel("test", "taskB", func() error {
		return nil
	})

	taskB.DependsOn(taskA, 50*time.Millisecond)

	tasks := []LabeledTask{taskB, taskA}
	err := ExecuteLabeled(tasks, 2, AlwaysReport())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for")
}

func Test_DependsOn_DependencyError_NoNotify(t *testing.T) {
	taskA := TaskWithLabel("test", "taskA", func() error {
		return fmt.Errorf("taskA failed")
	})

	taskB := TaskWithLabel("test", "taskB", func() error {
		return nil
	})

	taskB.DependsOn(taskA, 100*time.Millisecond)

	tasks := []LabeledTask{taskB, taskA}
	err := ExecuteLabeled(tasks, 2, AlwaysReport())
	require.Error(t, err)
	// taskA fails, so its notifier never fires, causing taskB to time out
	assert.Contains(t, err.Error(), "taskA failed")
}

func Test_DependsOn_MultipleDependents(t *testing.T) {
	var aCompleted atomic.Bool

	taskA := TaskWithLabel("test", "taskA", func() error {
		time.Sleep(50 * time.Millisecond)
		aCompleted.Store(true)
		return nil
	})

	taskB := TaskWithLabel("test", "taskB", func() error {
		assert.True(t, aCompleted.Load(), "taskB should run after taskA")
		return nil
	})

	taskC := TaskWithLabel("test", "taskC", func() error {
		assert.True(t, aCompleted.Load(), "taskC should run after taskA")
		return nil
	})

	taskB.DependsOn(taskA, 5*time.Second)
	taskC.DependsOn(taskA, 5*time.Second)

	tasks := []LabeledTask{taskB, taskC, taskA}
	err := ExecuteLabeled(tasks, 3, AlwaysReport())
	require.NoError(t, err)
}

func Test_DependsOn_Chain(t *testing.T) {
	var order []string
	orderCh := make(chan string, 3)

	taskA := TaskWithLabel("test", "taskA", func() error {
		time.Sleep(30 * time.Millisecond)
		orderCh <- "A"
		return nil
	})

	taskB := TaskWithLabel("test", "taskB", func() error {
		time.Sleep(30 * time.Millisecond)
		orderCh <- "B"
		return nil
	})

	taskC := TaskWithLabel("test", "taskC", func() error {
		orderCh <- "C"
		return nil
	})

	taskB.DependsOn(taskA, 5*time.Second)
	taskC.DependsOn(taskB, 5*time.Second)

	tasks := []LabeledTask{taskC, taskB, taskA}
	err := ExecuteLabeled(tasks, 3, AlwaysReport())
	require.NoError(t, err)

	close(orderCh)
	for v := range orderCh {
		order = append(order, v)
	}

	require.Equal(t, []string{"A", "B", "C"}, order)
}

func Test_GetNotifier_Idempotent(t *testing.T) {
	task := TaskWithLabel("test", "task", func() error {
		return nil
	})

	ch1 := getNotifier(task)
	ch2 := getNotifier(task)

	assert.Equal(t, ch1, ch2, "getNotifier should return the same channel on repeated calls")
}

func Test_WrapTask(t *testing.T) {
	var wrappedCalled bool

	task := TaskWithLabel("test", "task", func() error {
		return nil
	})

	task.WrapTask(func(inner Executable) Executable {
		return &testWrapper{inner: inner, called: &wrappedCalled}
	})

	err := task.Execute()
	require.NoError(t, err)
	assert.True(t, wrappedCalled)
}

type testWrapper struct {
	inner  Executable
	called *bool
}

func (w *testWrapper) Execute(task LabeledTask) error {
	*w.called = true
	return w.inner.Execute(task)
}
