package fun

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestTypes(t *testing.T) {
	suite.Run(t, new(TypesSuite))
}

type TypesSuite struct {
	suite.Suite
}

func (s *TypesSuite) TestConfig() {
	config := NewConfig[int]().
		Set("a", 1).
		Set("d", 4).
		Set("c", 3).
		Set("e", 5).
		Set("f", 6).
		Set("b", 2)

	var results []Entry[string, int]
	for r := range config.IteratorOrdered() {
		results = append(results, r)
	}
	s.Equal(
		[]Entry[string, int]{
			{"a", 1},
			{"b", 2},
			{"c", 3},
			{"d", 4},
			{"e", 5},
			{"f", 6},
		},
		results,
	)
}
