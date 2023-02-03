package build

import (
	"fmt"
	"testing"

	"github.com/itura/fun/pkg/build/mocks"
	"github.com/stretchr/testify/assert"
)

func TestApplySideEffectsRunsCommands(t *testing.T) {
	runner := new(mocks.CommandRunner)
	sideEffects := SideEffects{
		Commands: []Command{
			{
				Name: "name",
				Arguments: []string{
					"arg1",
					"arg2",
				},
			},
			{
				Name: "name",
				Arguments: []string{
					"arg3",
					"arg4",
				},
			},
		},
	}

	runner.On("Run", "name", "arg1", "arg2").Return(nil)
	runner.On("Run", "name", "arg3", "arg4").Return(nil)

	err := sideEffects.Apply(runner)
	assert.Nil(t, err)
	runner.AssertExpectations(t)

}

func TestApplySideEffectsReturnsFirstError(t *testing.T) {
	runner := new(mocks.CommandRunner)
	sideEffects := SideEffects{
		Commands: []Command{
			{
				Name: "name",
				Arguments: []string{
					"arg1",
					"arg2",
				},
			},
			{
				Name: "name",
				Arguments: []string{
					"arg3",
					"arg4",
				},
			},
		},
	}

	expectedErr := fmt.Errorf("Failed to run command")

	runner.On("Run", "name", "arg1", "arg2").Return(expectedErr)

	returnedErr := sideEffects.Apply(runner)

	assert.Equal(t, expectedErr, returnedErr)
	runner.AssertExpectations(t)

}
