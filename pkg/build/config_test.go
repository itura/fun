package build

import (
	"fmt"
	"github.com/itura/fun/pkg/fun"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	builder := NewTestBuilder()
	cases := []struct {
		args     ActionArgs
		name     string
		expected PipelineConfig
	}{
		{
			name:     "ValidPipelineConfig",
			args:     TestArgs("test_fixtures/valid_pipeline_config.yaml"),
			expected: ValidPipelineConfig(builder),
		},
		{
			name: "InvalidSecretName",
			args: TestArgs("test_fixtures/invalid_secret_name.yaml"),
			expected: FailedParse("My Build", NewValidationErrors("applications").
				PutChild(NewValidationErrors("db").
					PutChild(NewValidationErrors("secrets").
						Put("postgresql.auth.postgresPassword", fmt.Errorf("secret 'beepboop' not configured in any secretProvider")),
					),
				),
			),
		}, {
			name: "InvalidSecretProvider",
			args: TestArgs("test_fixtures/invalid_secret_provider.yaml"),
			expected: FailedParse("My Build", NewValidationErrors("").
				PutChild(NewValidationErrors("resources").
					PutChild(NewValidationErrors("secretProviders").
						PutChild(NewValidationErrors("0").
							Put("id", eMissingRequiredField).
							Put("secretNames", eMissingRequiredField),
						).
						PutChild(NewValidationErrors("1").
							Put("secretNames", eMissingRequiredField).
							PutChild(NewValidationErrors("config").
								Put("project", eMissingRequiredField)),
						).
						PutChild(NewValidationErrors("2").
							Put("type", eMissingRequiredField),
						).
						PutChild(NewValidationErrors("3").
							Put("secretNames", eMissingRequiredField).
							Put("config", eMissingRequiredField),
						),
					),
				),
			),
		},
		{
			name:     "InvalidSecretProviderType",
			args:     TestArgs("test_fixtures/invalid_secret_provider_type.yaml"),
			expected: FailedParse("", SecretProviderTypeEnum.InvalidEnumValue("aws")),
		},
		{
			name: "InvalidCloudProvider",
			args: TestArgs("test_fixtures/invalid_cloud_provider.yaml"),
			expected: FailedParse("My Build", NewValidationErrors("").
				PutChild(NewValidationErrors("resources").
					Put("cloudProvider", eMissingRequiredField).
					Put("kubernetesCluster", eMissingRequiredField).
					PutChild(NewValidationErrors("artifactRepository").
						Put("host", eMissingRequiredField)),
				)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseConfig(tc.args, NewAlwaysChanged())
			assert.Equal(t, tc.expected.BuildName, result.BuildName)
			assert.Equal(t, tc.expected.Artifacts, result.Artifacts)
			assert.Equal(t, tc.expected.Applications, result.Applications)
			fmt.Println(result.Error)
			assert.Equal(t, tc.expected.Error, result.Error)
		})
	}
}

func TestGithubActionsGeneration(t *testing.T) {
	t.Skip()
	result, err := ReadConfigForGeneration("test_fixtures/valid_pipeline_config.yaml", "???")
	assert.Nil(t, err)
	//assert.Equal(t, "yeehaw", result)
	_ = result.WriteYaml("test_fixtures/yeehaw.yaml")
}

func TestGithubActionsGeneration1(t *testing.T) {
	cmd := "github.com/itura/fun/cmd/build@v0.1.23"
	configPath := "pipeline.yaml"
	cases := []struct {
		name     string
		config   PipelineConfigRaw
		expected GitHubActionsWorkflow
	}{
		{
			name: "it generates a valid workflow",
			config: PipelineConfigRaw{
				Name:      "my build",
				Resources: ValidResources(),
				Artifacts: []ArtifactConfig{
					{Id: "api", Path: "pkg/api"},
				},
				Applications: []ApplicationConfig{{
					Id:   "infra",
					Type: applicationTypeTerraform,
					Path: "tf/main",
				}, {
					Id:           "website",
					Type:         applicationTypeHelm,
					Path:         "helm/website",
					Namespace:    "postgres",
					Artifacts:    []string{"api"},
					Dependencies: []string{"infra"},
					Values: []RuntimeArg{
						{Key: "domain", Value: "http://yeehaw.com"},
					},
					Secrets: []SecretConfig{
						{Key: "postgres.password", SecretName: "pg-password"},
						{Key: "postgres.adminPassword", SecretName: "pg-admin-password"},
					},
				}},
			},
			expected: NewGitHubActionsWorkflow("my build").
				SetJob("build-api", NewGitHubActionsJob("Build api").
					AddSteps(
						CheckoutRepoStep(),
						SetupGoStep(),
						GcpAuthStep("${{ secrets.WORKLOAD_IDENTITY_PROVIDER }}", "${{ secrets.SERVICE_ACCOUNT }}"),
						ConfigureGcloudCliStep(),
						ConfigureGcloudDockerStep("us"),
						BuildArtifactStep("api", configPath, cmd),
					),
				).
				SetJob("deploy-infra", NewGitHubActionsJob("Deploy infra").
					AddSteps(
						CheckoutRepoStep(),
						SetupGoStep(),
						GcpAuthStep("${{ secrets.WORKLOAD_IDENTITY_PROVIDER }}", "${{ secrets.SERVICE_ACCOUNT }}"),
						GetSetupTerraformStep(),
						GetDeployStep("infra", nil, GetDeployRunCommand("infra", cmd, configPath)),
					),
				).
				SetJob("deploy-website", NewGitHubActionsJob("Deploy website").
					AddNeeds("build-api", "deploy-infra").
					AddSteps(
						CheckoutRepoStep(),
						SetupGoStep(),
						GcpAuthStep("${{ secrets.WORKLOAD_IDENTITY_PROVIDER }}", "${{ secrets.SERVICE_ACCOUNT }}"),
						FetchGcpSecretsStep("gcp-cool-proj", "cool-proj", "pg-admin-password"),
						GetSetupGkeStep(ClusterConfig{
							Name:     "cluster",
							Location: "new zealand",
							Type:     "gke",
						}),
						GetSetupHelmStep(),
						GetDeployStep(
							"website",
							[]RuntimeArg{
								{Key: "domain", Value: "http://yeehaw.com"},
								{Key: "postgres.password", Value: "${{ secrets.pg-password }}"},
								{Key: "postgres.adminPassword", Value: "${{ steps.secrets-gcp-cool-proj.outputs.pg-admin-password }}"},
							},
							GetDeployRunCommand("website", cmd, configPath),
						),
					),
				),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseConfigForGeneration(tc.config, configPath, cmd)
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCloudProviderValidations(t *testing.T) {
	cp := CloudProviderConfig{
		Type: cloudProviderTypeGcp,
		Config: fun.NewConfig[string]().
			Set("serviceAccount", "yeehaw@yahoo.com").
			Set("workloadIdentityProvider", "it me"),
	}

	errs := cp.Validate("cloudProvider")
	assert.Equal(t, false, errs.IsPresent())
	assert.Equal(t,
		NewValidationErrors("cloudProvider"),
		errs,
	)

	cp = CloudProviderConfig{
		Type:   cloudProviderTypeGcp,
		Config: fun.NewConfig[string](),
	}

	errs = cp.Validate("cloudProvider")
	assert.Equal(t, true, errs.IsPresent())
	assert.Equal(t,
		NewValidationErrors("cloudProvider").
			PutChild(NewValidationErrors("config").
				Put("serviceAccount", CloudProviderMissingField("gcp")).
				Put("workloadIdentityProvider", CloudProviderMissingField("gcp"))),
		errs,
	)

	cp = CloudProviderConfig{
		Type: cloudProviderTypeGcp,
		Config: fun.NewConfig[string]().
			Set("serviceAccount", "yeehaw@yahoo.com"),
	}
	errs = cp.Validate("cloudProvider")
	assert.Equal(t, true, errs.IsPresent())
	assert.Equal(t,
		NewValidationErrors("cloudProvider").
			PutChild(NewValidationErrors("config").
				Put("workloadIdentityProvider", CloudProviderMissingField("gcp"))),
		errs,
	)
}

func TestResourcesValidation(t *testing.T) {
	resources := ValidResources()

	errs := resources.Validate("resources")
	assert.Equal(t, false, errs.IsPresent())

	resources = Resources{
		SecretProviders: SecretProviderConfigs{
			SecretProviderConfig{
				Id:   "gcp-cool-proj",
				Type: secretProviderTypeGcp,
				Config: fun.Config[string]{
					"project": "cool-proj",
				},
			},
			SecretProviderConfig{},
		},
		CloudProvider: CloudProviderConfig{
			Type: cloudProviderTypeGcp,
			Config: fun.NewConfig[string]().
				Set("serviceAccount", "yeehaw@yahoo.com"),
		},
		ArtifactRepository: ArtifactRepository{
			Host: "us",
			Name: "repo",
		},
		KubernetesCluster: ClusterConfig{
			Name:     "cluster",
			Location: "new zealand",
		},
	}

	errs = resources.Validate("resources")
	assert.Equal(t, true, errs.IsPresent())
	//assert.Equal(t,
	//	NewValidationErrors("secretProviders"),
	//	errs,
	//)
	fmt.Println(errs.Error())
}

func TestValidateTags(t *testing.T) {

}
