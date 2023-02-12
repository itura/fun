package fun

import (
	"container/heap"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/constraints"
)

type JSON map[string]interface{}

func (j JSON) Merge(data ...JSON) JSON {
	for _, d := range data {
		for k, v := range d {
			j[k] = v
		}
	}
	return j
}

func (j JSON) Marshal() ([]byte, error) {
	return json.Marshal(j)
}

//func (j JSON) Get(keys ...string) (interface{}, bool) {
//	var result interface{}
//	current := j
//	for _, key := range keys {
//		next, present := current[key]
//		if !present {
//			return nil, false
//		}
//	}
//}

type Error string

func (e Error) Error() string {
	return string(e)
}

type Resource interface {
	Apply(Router gin.IRouter)
}

type Config[T any] map[string]T

func NewConfig[T any]() Config[T] {
	return Config[T]{}
}

func (c Config[T]) Set(key string, value T) Config[T] {
	c[key] = value
	return c
}

func (c Config[T]) SetAll(config Config[T]) Config[T] {
	for k, v := range config {
		c[k] = v
	}
	return c
}

func (c Config[T]) ForEach(fn func(string, T)) Config[T] {
	for k, v := range c {
		fn(k, v)
	}
	return c
}

func (c Config[T]) Iterator() <-chan Entry[string, T] {
	ch := make(chan Entry[string, T], len(c))
	go func() {
		for k, v := range c {
			ch <- NewEntry(k, v)
		}
		close(ch)
	}()
	return ch
}

func (c Config[T]) IteratorOrdered() <-chan Entry[string, T] {
	return MapEntriesOrdered(c)
}

func MapEntriesOrdered[K constraints.Ordered, V any](c map[K]V) <-chan Entry[K, V] {
	ch := make(chan Entry[K, V], len(c))
	go func() {
		h := &Heap[string]{}
		heap.Init(h)
		for key, _ := range c {
			heap.Push(h, key)
		}
		for i := 0; i < len(c); i++ {
			key := heap.Pop(h).(K)
			ch <- NewEntry(key, c[key])
		}
		close(ch)
	}()
	return ch
}

type MapIterator struct {
}

type Entry[K any, V any] struct {
	K K
	V V
}

func NewEntry[K any, V any](k K, v V) Entry[K, V] {
	return Entry[K, V]{K: k, V: v}
}

func (p Entry[K, V]) Get() (K, V) {
	return p.K, p.V
}

type Heap[T constraints.Ordered] []T

func (h Heap[T]) Len() int           { return len(h) }
func (h Heap[T]) Less(i, j int) bool { return h[i] < h[j] }

func (h Heap[T]) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *Heap[T]) Push(data any) {
	*h = append(*h, data.(T))
}

func (h *Heap[T]) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
