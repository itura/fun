package build

import (
	"reflect"
	"testing"
)

type InputArguments struct {
	ProjectId   string
	CurrentSha  string
	PreviousSha string
	Force       bool
}

type ParseConfigOutputs struct {
	Artifacts    map[string]Artifact
	Applications map[string]Application
	BuildName    string
	Error        error
}

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
				Artifacts: map[string]Artifact{
					"api": {
						Id:              "api",
						Path:            "packages/api",
						Project:         "projectId",
						Repository:      "us-central1-docker.pkg.dev/projectId/repo-name",
						Host:            "us-central1-docker.pkg.dev",
						CurrentSha:      "currentSha",
						Type:            ArtifactType("app"),
						hasDependencies: false,
						hasChanged:      true,
					},
				},
				Applications: map[string]Application{
					"db": {
						Id:         "db",
						Path:       "helm/db",
						ProjectId:  "projectId",
						Repository: "us-central1-docker.pkg.dev/projectId/repo-name",
						KubernetesCluster: ClusterConfig{
							Name:     "cluster-name",
							Location: "uscentral1",
						},
						CurrentSha: "currentSha",
						Namespace:  "app-namespace",
						Values: []HelmValue{
							{
								Key:   "postgresql.dbName",
								Value: "my-db",
							},
						},
						Upstreams: nil,
						Type:      ApplicationType("helm"),
						Secrets: []HelmSecretValue{
							{
								HelmKey:    "postgresql.auth.password",
								SecretName: "pg-password",
								Provider: SecretProvider{
									Type:   "github-actions",
									Config: nil,
								},
							},
						},
						hasDependencies: false,
						hasChanged:      true,
					},
				},
				BuildName: "My Build",
				Error:     nil,
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
			name: "InvalidSecretProvider",
			expectedOutputs: ParseConfigOutputs{
				Artifacts:    nil,
				Applications: nil,
				BuildName:    "",
				Error:        &InvalidSecretProvider{},
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

			if err != tc.expectedOutputs.Error {
				t.Fatalf("Expected error %v but got %v", tc.expectedOutputs.Error, err)
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
