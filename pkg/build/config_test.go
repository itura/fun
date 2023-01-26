package build

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseConfig(t *testing.T) {
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
				getValidArtifacts(),
				getValidApplications(),
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
			assert.Equal(t, tc.expected, result)
		})
	}
}
