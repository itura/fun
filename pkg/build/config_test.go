package build

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseConfig(t *testing.T) {
	builder := NewTestBuilder("projectId", "currentSha")
	cases := []struct {
		args     ActionArgs
		name     string
		expected ParsedConfig
	}{
		{
			name: "ValidPipelineConfig",
			args: TestArgs("test_fixtures/pipeline_config_pass.yaml"),
			expected: SuccessfulParse(
				"My Build",
				map[string]Artifact{
					"api":    builder.Artifact("api", "packages/api"),
					"client": builder.Artifact("client", "packages/client"),
				},
				map[string]Application{
					"db": builder.Application("db", "helm/db").
						AddValue("postgresql.dbName", "my-db").
						SetSecret("postgresql.auth.password", "princess-pup", "pg-password").
						SetSecret("postgresql.auth.username", "github", "pg-username"),
				},
			),
		},
		{
			name:     "MissingSecretProvider",
			args:     TestArgs("test_fixtures/missing_secret_provider.yaml"),
			expected: FailedParse("My Build", MissingSecretProvider{}),
		},
		{
			name:     "InvalidSecretProviderType",
			args:     TestArgs("test_fixtures/invalid_secret_provider.yaml"),
			expected: FailedParse("My Build", InvalidSecretProviderType{GivenType: "aws"}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseConfig(tc.args, "previousSha")
			assert.Equal(t, tc.expected.BuildName, result.BuildName)
			assert.Equal(t, tc.expected.Artifacts, result.Artifacts)
			assert.Equal(t, tc.expected.Applications, result.Applications)
			assert.Equal(t, tc.expected.Error, result.Error)
		})
	}
}
