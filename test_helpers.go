package fun

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"net"
	"net/http"
)

type ResourceTestSuite struct {
	suite.Suite
	server   *gin.Engine
	listener net.Listener
	Client   *RestClient
}

func (s *ResourceTestSuite) SetupServer(resources ...Resource) {
	s.server = gin.Default()
	for _, resource := range resources {
		resource.Apply(s.server)
	}

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		s.FailNow("couldn't set up resource test", err)
	}
	s.listener = listener

	go func() {
		_ = http.Serve(s.listener, s.server)
	}()

	host := fmt.Sprintf("http://%s", s.listener.Addr().String())
	s.Client = NewRestClient(host)
}

func (s *ResourceTestSuite) TearDownServer() {
	s.listener.Close()
}
