package build

import (
	"fmt"
	"os"

	"github.com/itura/fun/pkg/fun"
	"gopkg.in/yaml.v3"
)

func parseConfig(args ActionArgs, previousSha string) PipelineConfig {
	config, err := readFile(args.ConfigPath)
	if err != nil {
		return FailedParse("", err)
	}

	validationErrors := config.Validate("")
	if validationErrors.IsPresent() {
		return FailedParse(config.Name, validationErrors)
	}

	artifactRepository := fmt.Sprintf(
		"%s/%s",
		config.Resources.ArtifactRepository.Host,
		config.Resources.ArtifactRepository.Name,
	)
	artifacts := CreateArtifacts(args, previousSha, config, artifactRepository)
	applications, err := CreateApplications(args, previousSha, config, artifacts, artifactRepository)
	if err != nil {
		return FailedParse(config.Name, err)
	}

	return SuccessfulParse(config.Name, artifacts, applications)
}

func readFile(configPath string) (PipelineConfigRaw, error) {
	dat, err := os.ReadFile(configPath)
	if err != nil {
		return PipelineConfigRaw{}, err
	}

	var config PipelineConfigRaw
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return PipelineConfigRaw{}, err
	}

	return config, nil
}

type PipelineConfigRaw struct {
	Name      string    `validate:"required"`
	Resources Resources `validate:"required"`
	Artifacts []struct {
		Id           string
		Path         string
		Dependencies []string
		Type         ArtifactType
	}
	Applications []struct {
		Id           string
		Path         string
		Namespace    string
		Artifacts    []string
		Values       []HelmValue
		Secrets      []SecretConfig
		Dependencies []string
		Type         ApplicationType
	}
}

func (p PipelineConfigRaw) Validate(key string) ValidationErrors {
	return NewValidationErrors(key).Validate(p)
}

type Resources struct {
	ArtifactRepository ArtifactRepository  `yaml:"artifactRepository" validate:"required"`
	KubernetesCluster  ClusterConfig       `yaml:"kubernetesCluster"  validate:"required"`
	SecretProviders    SecretProviders     `yaml:"secretProviders"    validate:"required"`
	CloudProvider      CloudProviderConfig `yaml:"cloudProvider"      validate:"required"`
}

func (r Resources) Validate(key string) ValidationErrors {
	return NewValidationErrors(key).Validate(r)
}

type ClusterConfig struct {
	Name     string `validate:"required"`
	Location string `validate:"required"`
}

func (c ClusterConfig) Validate(key string) ValidationErrors {
	return NewValidationErrors(key).Validate(c)
}

type ArtifactRepository struct {
	Host string `validate:"required"`
	Name string `validate:"required"`
}

func (a ArtifactRepository) Validate(key string) ValidationErrors {
	return NewValidationErrors(key).Validate(a)
}

type CloudProviderConfig struct {
	Type   CloudProviderType
	Config map[string]string
}

func (c CloudProviderConfig) Impl() CloudProvider {
	switch c.Type {
	case cloudProviderTypeGcp:
		return GCP{c.Config}
	default:
		return nil
	}
}

func (c CloudProviderConfig) Validate(key string) ValidationErrors {
	return NewValidationErrors(key).
		PutChild(c.Impl().Validate("config"))
}

type SecretProvider struct {
	Type   SecretProviderType `validate:"required"`
	Config map[string]string
}

type SecretConfig struct {
	HelmKey    string "yaml:\"helmKey\""
	SecretName string "yaml:\"secretName\""
	Provider   string
}

func (s SecretProvider) Validate(key string) ValidationErrors {
	return NewValidationErrors(key).ValidateTags(s)
}

type SecretProviders fun.Config[SecretProvider]

func (s SecretProviders) Validate(key string) ValidationErrors {
	errs := NewValidationErrors(key)
	for k, provider := range s {
		errs = errs.PutChild(NewValidationErrors(k).Validate(provider))
	}
	return errs
}
