package fun

import "github.com/gin-gonic/gin"

type JSON map[string]interface{}
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
