package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
					"db": PostgresHelmChart(builder),
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
		{
			name:     "InvalidCloudProvider",
			args:     TestArgs("test_fixtures/invalid_cloud_provider.yaml"),
			expected: FailedParse("My Build", InvalidCloudProvider{Message: "Missing/Unknown"}),
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
