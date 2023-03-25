package result

type Result[T any] struct {
	Value T
	Err   error
}

func Success[T any](value T) Result[T] {
	return Result[T]{Value: value}
}

func Failure[T any](err error) Result[T] {
	return Result[T]{Err: err}
}

func (r Result[T]) Ok() bool {
	return r.Err == nil
}

func (r Result[T]) Map(fs ...func(T) Result[T]) Result[T] {
	current := r
	for _, f := range fs {
		if current.Ok() {
			current = f(current.Value)
		}
	}
	return current
}

func Unit[X comparable](x X) Result[X] {
	return Success(x)
}

func Lift[X, Y comparable](f func(X) (Y, error)) func(X) Result[Y] {
	return func(x X) Result[Y] {
		value, err := f(x)
		if err != nil {
			return Failure[Y](err)
		}
		return Unit(value)
	}
}

type Fn[X, Y comparable] func(X) Result[Y]

func Bind[X, Y comparable](r Result[X], f Fn[X, Y]) Result[Y] {
	if r.Ok() {
		return f(r.Value)
	} else {
		return Result[Y]{Err: r.Err}
	}
}

func Compose[X, Y, Z comparable](f1 Fn[X, Y], f2 Fn[Y, Z]) Fn[X, Z] {
	return func(x X) Result[Z] {
		xy := Bind(Unit(x), f1)
		return Bind(xy, f2)
	}
}

func Compose1[X, Y, Z, A comparable](f1 Fn[X, Y], f2 Fn[Y, Z], f3 Fn[Z, A]) Fn[X, A] {
	return func(x X) Result[A] {
		xy := Bind(Unit(x), f1)
		yz := Bind(xy, f2)
		return Bind(yz, f3)
	}
}

func Compose2[X, Y, Z, A, B comparable](f1 Fn[X, Y], f2 Fn[Y, Z], f3 Fn[Z, A], f4 Fn[A, B]) Fn[X, B] {
	return func(x X) Result[B] {
		xy := Bind(Unit(x), f1)
		yz := Bind(xy, f2)
		za := Bind(yz, f3)
		return Bind(za, f4)
	}
}
