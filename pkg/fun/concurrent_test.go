package fun

import (
	"fmt"
	"github.com/itura/fun/pkg/fun/result"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWorkers_Listen(t *testing.T) {
	workers := NewWorkers[string](3)
	workers.Submit(
		func() result.Result[string] { return result.Success("hi") },
		func() result.Result[string] { return result.Success("there") },
		func() result.Result[string] { return result.Success("my") },
		func() result.Result[string] { return result.Success("friendo") },
	)

	ch := workers.Listen()
	var results []result.Result[string]
	for i := 0; i < 4; i++ {
		results = append(results, <-ch)
	}
	assert.Equal(t,
		[]result.Result[string]{
			result.Success("hi"),
			result.Success("there"),
			result.Success("my"),
			result.Success("friendo"),
		},
		results,
	)

	ok := workers.Stop()
	assert.True(t, ok)
}

func TestWorkers_Collect(t *testing.T) {
	workers := NewWorkers[string](3)
	workers.Submit(
		func() result.Result[string] { return result.Success("hi") },
		func() result.Result[string] { return result.Success("there") },
		func() result.Result[string] { return result.Success("my") },
		func() result.Result[string] { return result.Success("friendo") },
	)

	results := workers.Collect()
	assert.Equal(t,
		[]result.Result[string]{
			result.Success("hi"),
			result.Success("there"),
			result.Success("my"),
			result.Success("friendo"),
		},
		results,
	)

	ok := workers.Stop()
	assert.True(t, ok)
}

func TestWorkers_async(t *testing.T) {
	bc := NewBroadcastChannel(3)
	workers := NewWorkers[string](3)
	workers.Submit(
		blockingTask(0, bc, func() result.Result[string] { return result.Success("hi") }),
	)

	ch := workers.Listen()
	select {
	case _ = <-ch:
		assert.FailNow(t, "no tasks should have completed")
	case <-time.After(100 * time.Millisecond):
		bc.Notify(0)
		assert.Equal(t, result.Success("hi"), <-ch)
	}

	workers.Submit(
		blockingTask(1, bc, func() result.Result[string] { return result.Success("there") }),
		blockingTask(2, bc, func() result.Result[string] { return result.Success("my") }),
		blockingTask(3, bc, func() result.Result[string] { return result.Success("friendo") }),
		blockingTask(4, bc, func() result.Result[string] { return result.Success("!!") }),
	)

	bc.Notify(4)
	select {
	case _ = <-ch:
		assert.FailNow(t, "no tasks should have completed")
	case <-time.After(100 * time.Millisecond):
		bc.Notify(1)
		assert.Equal(t, result.Success("there"), <-ch)
		assert.Equal(t, result.Success("!!"), <-ch) // task 4 completes as soon as a worker picks it up
	}

	ok := workers.Stop()
	assert.True(t, ok)
}

func blockingTask[T comparable](id int, bc *BroadcastChannel, task Task[T]) Task[T] {
	finish := bc.Listen()
	return func() result.Result[T] {
		for t := range finish {
			switch v := t.(type) {
			case int:
				if v == id {
					return task()
				}
			}
		}
		return result.Failure[T](fmt.Errorf("not invoked"))
	}
}
