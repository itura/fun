package fun

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

type ClientSuite struct {
	suite.Suite
}

func (s *ClientSuite) TestConfig() {
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

func (s *ClientSuite) TestHttpParamsQueryParams() {
	params := NewHttpParams()
	s.Empty(params.Header)
	s.Empty(params.Query)

	s.Equal(
		Config[string]{
			"yee":  "haw",
			"beep": "boop",
		},
		params.Query.
			Set("yee", "haw").
			Set("beep", "boop"),
	)

	s.Equal(
		&HttpParams{
			Query: Config[string]{
				"yee":  "haw",
				"beep": "boop!",
				"hoo":  "wee",
			},
			Header: Config[string]{},
		},
		params.SetQuery(
			params.Query.
				Set("beep", "boop!").
				Set("hoo", "wee"),
		),
	)
}
