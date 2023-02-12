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
			name:     "MissingSecretProvider",
			args:     TestArgs("test_fixtures/missing_secret_provider.yaml"),
			expected: FailedParse("My Build", MissingSecretProvider{}),
		},
		{
			name:     "InvalidSecretProviderType",
			args:     TestArgs("test_fixtures/invalid_secret_provider.yaml"),
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
			result := parseConfig(tc.args, "previousSha")
			assert.Equal(t, tc.expected.BuildName, result.BuildName)
			assert.Equal(t, tc.expected.Artifacts, result.Artifacts)
			assert.Equal(t, tc.expected.Applications, result.Applications)
			assert.Equal(t, tc.expected.Error, result.Error)
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

func TestSecretProvidersValidation(t *testing.T) {
	sp := SecretProviders{
		"one": SecretProviderRaw{
			Type:   secretProviderTypeGithub,
			Config: nil,
		},
		"two": SecretProviderRaw{
			Type: secretProviderTypeGcp,
			Config: fun.Config[string]{
				"project": "cool-proj",
			},
		},
	}

	errs := sp.Validate("secretProviders")
	assert.Equal(t, false, errs.IsPresent())
}

func TestResourcesValidation(t *testing.T) {
	resources := Resources{
		SecretProviders: SecretProviders{
			"one": SecretProviderRaw{
				Type:   secretProviderTypeGithub,
				Config: nil,
			},
			"two": SecretProviderRaw{
				Type: secretProviderTypeGcp,
				Config: fun.Config[string]{
					"project": "cool-proj",
				},
			},
		},
		CloudProvider: CloudProviderConfig{
			Type: cloudProviderTypeGcp,
			Config: fun.NewConfig[string]().
				Set("serviceAccount", "yeehaw@yahoo.com").
				Set("workloadIdentityProvider", "it me"),
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

	errs := resources.Validate("resources")
	assert.Equal(t, false, errs.IsPresent())

	resources = Resources{
		SecretProviders: SecretProviders{
			"two": SecretProviderRaw{
				Type: secretProviderTypeGcp,
				Config: fun.Config[string]{
					"project": "cool-proj",
				},
			},
			"four": SecretProviderRaw{},
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
