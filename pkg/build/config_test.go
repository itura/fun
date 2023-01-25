package build

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseConfig(t *testing.T) {
	cases := []struct {
		fixture         string
		arguments       InputArguments
		name            string
		expectedOutputs ParseConfigOutputs
	}{
		{
			fixture: "test_fixtures/pipeline_config_pass.yaml",
			arguments: InputArguments{
				ProjectId:   "projectId",
				CurrentSha:  "currentSha",
				PreviousSha: "previousSha",
				Force:       false,
			},
			name: "ValidPipelineConfig",
			expectedOutputs: ParseConfigOutputs{
				Artifacts:    getValidArtifacts(),
				Applications: getValidApplications(),
				BuildName:    "My Build",
				Error:        nil,
			},
		},
		{
			fixture: "test_fixtures/missing_secret_provider.yaml",
			arguments: InputArguments{
				ProjectId:   "projectId",
				CurrentSha:  "currentSha",
				PreviousSha: "previousSha",
				Force:       false,
			},
			name: "MissingSecretProvider",
			expectedOutputs: ParseConfigOutputs{
				Artifacts:    nil,
				Applications: nil,
				BuildName:    "",
				Error:        MissingSecretProvider{},
			},
		},
		{
			fixture: "test_fixtures/invalid_secret_provider.yaml",
			arguments: InputArguments{
				ProjectId:   "projectId",
				CurrentSha:  "currentSha",
				PreviousSha: "previousSha",
				Force:       false,
			},
			name: "InvalidSecretProviderType",
			expectedOutputs: ParseConfigOutputs{
				Artifacts:    nil,
				Applications: nil,
				BuildName:    "",
				Error:        InvalidSecretProviderType{GivenType: "aws"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			artifacts, applications, name, err := parseConfig(tc.fixture,
				tc.arguments.ProjectId,
				tc.arguments.CurrentSha,
				tc.arguments.PreviousSha,
				tc.arguments.Force,
			)

			if !errors.Is(err, tc.expectedOutputs.Error) {
				t.Fatalf("Expected error '%v' but got '%v'", tc.expectedOutputs.Error, err)
			}

			if name != tc.expectedOutputs.BuildName {
				t.Fatalf("Expected build name to be %s but got %s instead", tc.expectedOutputs.BuildName, name)
			}

			if !reflect.DeepEqual(artifacts, tc.expectedOutputs.Artifacts) {
				t.Fatalf("Expected artifacts to be \n%v but got \n%v", tc.expectedOutputs.Artifacts, artifacts)
			}

			if !reflect.DeepEqual(applications, tc.expectedOutputs.Applications) {
				t.Fatalf("Expected applications to be \n%v but got \n%v", tc.expectedOutputs.Applications, applications)
			}

		})
	}
}
