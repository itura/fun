package build

import (
	"github.com/itura/fun/pkg/fun"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestCd(t *testing.T) {
	suite.Run(t, new(CdSuite))
}

type CdSuite struct {
	suite.Suite
}

func (s *CdSuite) TestDependency() {
	config, err := readFile("test_fixtures/valid_pipeline_config.yaml")
	s.Nil(err)

	deps := ParseDependencies(config)
	s.Equal(
		Dependencies{
			deps: fun.NewConfig[Dependency]().
				Set("client", NewArtifactDependency("client", "packages/client")).
				Set("db", NewApplicationDependency("db", "helm/db", "infra")).
				Set("infra", NewApplicationDependency("infra", "tf/main")).
				Set("website", NewApplicationDependency("website", "helm/website", "client", "api", "infra", "db")).
				Set("api", NewArtifactDependency("api", "packages/api")),
		},
		deps,
	)

	s.Equal([]string(nil), deps.GetUpstreamJobIds("client"))
	s.Equal([]string{"deploy-infra"}, deps.GetUpstreamJobIds("db"))
	s.Equal([]string{"build-client", "build-api", "deploy-infra", "deploy-db"}, deps.GetUpstreamJobIds("website"))
	s.Equal([]string(nil), deps.GetUpstreamJobIds("infra"))
	s.Equal([]string(nil), deps.GetUpstreamJobIds("api"))

	s.Equal([]string{"packages/client"}, deps.GetAllPaths("client"))
	s.Equal([]string{"helm/db", "tf/main"}, deps.GetAllPaths("db"))
	s.Equal([]string{"tf/main"}, deps.GetAllPaths("infra"))
	s.Equal([]string{"helm/website", "packages/client", "packages/api", "tf/main", "helm/db"}, deps.GetAllPaths("website"))
	s.Equal([]string{"packages/api"}, deps.GetAllPaths("api"))
}
