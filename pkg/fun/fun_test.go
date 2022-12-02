package fun

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestPickleResources(t *testing.T) {
	suite.Run(t, new(FunTestSuite))
}

type FunTestSuite struct {
	suite.Suite
}

func (s *FunTestSuite) TestConfig() {
	config := NewConfig[string]()
	s.Equal(
		Config[string]{
			"yee": "haw",
			"hoo": "wee",
		},
		config.
			Set("yee", "haw").
			Set("hoo", "wee"),
	)

	config.ForEach(func(k string, v string) {
		s.Contains([]string{"haw", "wee"}, v)
	})
}
