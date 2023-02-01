package fun

type Result[T any] struct {
	value T
	err   error
}

func Success[T any](value T) Result[T] {
	return Result[T]{value: value}
}
func Failure[T any](err error) Result[T] {
	return Result[T]{err: err}
}
