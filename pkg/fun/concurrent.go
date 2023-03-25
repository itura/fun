package fun

import "github.com/itura/fun/pkg/fun/result"

type Task[T any] func() result.Result[T]

type Workers[T any] struct {
	count     int
	taskCount int
	tasks     chan Task[T]
	results   chan result.Result[T]
}

func NewWorkers[T any](count int) *Workers[T] {
	tasks := make(chan Task[T], count)
	results := make(chan result.Result[T], count)

	for i := 0; i < count; i++ {
		go func() {
			for task := range tasks {
				results <- task()
			}
		}()
	}

	return &Workers[T]{
		count:   count,
		tasks:   tasks,
		results: results,
	}
}

func (w *Workers[T]) Stop() bool {
	if len(w.tasks) == 0 {
		close(w.tasks)
	} else {
		return false
	}

	if len(w.results) == 0 {
		close(w.results)
	} else {
		return false
	}

	return true
}

func (w *Workers[T]) Submit(tasks ...Task[T]) {
	for _, task := range tasks {
		w.taskCount += 1
		w.tasks <- task
	}
}

func (w *Workers[T]) Listen() <-chan result.Result[T] {
	return w.results
}

func (w *Workers[T]) Collect(count ...int) []result.Result[T] {
	var _count int
	if len(count) == 1 {
		_count = count[0]
	} else {
		_count = w.taskCount
	}

	var results []result.Result[T]
	for i := 0; i < _count; i++ {
		results = append(results, <-w.results)
	}
	return results
}

type BroadcastChannel struct {
	channels []chan interface{}
	buffer   int
}

func NewBroadcastChannel(buffer int) *BroadcastChannel {
	return &BroadcastChannel{buffer: buffer}
}

func (bc *BroadcastChannel) Listen() <-chan interface{} {
	channel := make(chan interface{}, bc.buffer)
	bc.channels = append(bc.channels, channel)
	return channel
}

func (bc *BroadcastChannel) Notify(msg interface{}) {
	for _, channel := range bc.channels {
		channel <- msg
	}
}

func (bc *BroadcastChannel) Close() {
	for _, channel := range bc.channels {
		close(channel)
	}
}
