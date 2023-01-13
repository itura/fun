package fun

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
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
