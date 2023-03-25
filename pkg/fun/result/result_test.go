package result

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type state struct {
	a string
	b int
}

func TestFunctor(t *testing.T) {
	result := Success(state{"hi", 1})
	assert.True(t, result.Ok())

	inc := func(value state) Result[state] {
		value.a += "!"
		value.b += 1
		return Success(value)
	}
	fail := func(value state) Result[state] {
		return Failure[state](fmt.Errorf("ah beans"))
	}
	result = result.Map(inc, inc)
	assert.True(t, result.Ok())
	assert.Equal(t, state{"hi!!", 3}, result.Value)

	result = result.Map(fail, inc)
	assert.False(t, result.Ok())
	assert.Equal(t, state{}, result.Value)
	assert.EqualError(t, result.Err, "ah beans")
}

func TestMonad(t *testing.T) {
	result := Success("hi")
	result1 := Bind(result, func(x string) Result[int] {
		return Success(len(x))
	})
	assert.True(t, result1.Ok())
	assert.Equal(t, 2, result1.Value)

	length := Lift(func(x string) (int, error) {
		l := len(x)
		if l > 5 {
			return 0, fmt.Errorf("i can't count that high")
		}
		return l, nil
	})
	lessThan := func(threshold int) func(x int) Result[bool] {
		return Lift(func(x int) (bool, error) {
			return x < threshold, nil
		})
	}
	toMessage := Lift(func(x bool) (string, error) {
		if x {
			return "yay", nil
		} else {
			return "", fmt.Errorf("ouch")
		}
	})

	// one fn
	result = Success("bye")
	result1 = Bind(result, length)
	assert.True(t, result1.Ok())
	assert.Equal(t, 3, result1.Value)

	result = Success("byeeeeeee")
	result1 = Bind(result, length)
	assert.False(t, result1.Ok())
	assert.EqualError(t, result1.Err, "i can't count that high")

	// two fn
	result = Success("bye")
	result2 := Bind(result, Compose(length, lessThan(4)))
	assert.True(t, result2.Ok())
	assert.Equal(t, true, result2.Value)

	result = Success("byeeeeeee")
	result2 = Bind(result, Compose(length, lessThan(4)))
	assert.False(t, result2.Ok())
	assert.EqualError(t, result2.Err, "i can't count that high")

	// three fn
	result = Success("bye")
	result3 := Bind(result, Compose1(length, lessThan(4), toMessage))
	assert.True(t, result3.Ok())
	assert.Equal(t, "yay", result3.Value)

	result = Success("bye")
	result3 = Bind(result, Compose1(length, lessThan(2), toMessage))
	assert.False(t, result3.Ok())
	assert.EqualError(t, result3.Err, "ouch")

	result = Success("byeeeeeee")
	result3 = Bind(result, Compose1(length, lessThan(2), toMessage))
	assert.False(t, result3.Ok())
	assert.EqualError(t, result3.Err, "i can't count that high")
}
