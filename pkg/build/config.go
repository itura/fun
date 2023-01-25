package build

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Preset string

var (
	presetGolang Preset = "golang"
)

type ArtifactType string

var (
	typeLibGo ArtifactType = "lib-go"
	typeAppGo ArtifactType = "app-go"
	typeApp   ArtifactType = "app"
	typeLib   ArtifactType = "lib"
)

type ApplicationType string

var (
	typeHelm      ApplicationType = "helm"
	typeTerraform ApplicationType = "terraform"
)

type ClusterConfig struct {
	Name     string
	Location string
}

type SecretProvider struct {
	Type   string
	Config map[string]string
}

type SecretConfig struct {
	HelmKey    string "yaml:\"helmKey\""
	SecretName string "yaml:\"secretName\""
	Provider   string
}

type PipelineConfig struct {
	Name      string
	Resources struct {
		ArtifactRepository struct {
			Host string
			Name string
		} `yaml:"artifactRepository"`

		KubernetesCluster ClusterConfig             `yaml:"kubernetesCluster"`
		SecretProviders   map[string]SecretProvider `yaml:"secretProviders"`
	}
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

func parseConfig(configPath string, projectId string, currentSha string, previousSha string, force bool) (map[string]Artifact, map[string]Application, string, error) {
	dat, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, "", err
	}

	var config PipelineConfig
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return nil, nil, "", err
	}

	var repository string = fmt.Sprintf("%s/%s/%s", config.Resources.ArtifactRepository.Host, projectId, config.Resources.ArtifactRepository.Name)
	var providerConfigs map[string]SecretProvider = config.Resources.SecretProviders

	artifacts := make(map[string]Artifact)
	for _, spec := range config.Artifacts {
		var cd ChangeDetection
		if force {
			cd = NewAlwaysChanged()
		} else {
			_cd := NewGitChangeDetection(previousSha).
				AddPaths(spec.Path)

			// todo make agnostic to ordering
			for _, id := range spec.Dependencies {
				_cd = _cd.AddPaths(artifacts[id].Path)
			}
			cd = _cd
		}

		artifacts[spec.Id] = Artifact{
			Type:            spec.Type,
			Id:              spec.Id,
			Path:            spec.Path,
			Project:         projectId,
			Repository:      repository,
			Host:            config.Resources.ArtifactRepository.Host,
			CurrentSha:      currentSha,
			hasDependencies: len(spec.Dependencies) > 0,
			hasChanged:      cd.HasChanged(),
		}
	}

	applications := make(map[string]Application)
	for _, spec := range config.Applications {
		var upstreams []Job
		var cd ChangeDetection
		if force {
			cd = NewAlwaysChanged()
		} else {
			_cd := NewGitChangeDetection(previousSha).
				AddPaths(spec.Path)

			// todo make agnostic to ordering
			for _, id := range spec.Artifacts {
				_cd = _cd.AddPaths(artifacts[id].Path)
				upstreams = append(upstreams, artifacts[id])
			}
			for _, id := range spec.Dependencies {
				_cd = _cd.AddPaths(applications[id].Path)
				upstreams = append(upstreams, applications[id])
			}
			cd = _cd
		}

		var secretConfigs = spec.Secrets

		helmSecretValues := make([]HelmSecretValue, len(secretConfigs))
		for i, secretConfig := range secretConfigs {
			provider, ok := providerConfigs[secretConfig.Provider]

			if !ok {
				return nil, nil, "", &InvalidSecretProvider{}
			}

			helmSecretValue := HelmSecretValue{
				HelmKey:    secretConfig.HelmKey,
				SecretName: secretConfig.SecretName,
				Provider:   provider,
			}

			helmSecretValues[i] = helmSecretValue
		}

		hasDependencies := len(spec.Dependencies) > 0 || len(spec.Artifacts) > 0
		applications[spec.Id] = Application{
			Type:              spec.Type,
			Id:                spec.Id,
			Path:              spec.Path,
			ProjectId:         projectId,
			Repository:        repository,
			CurrentSha:        currentSha,
			Namespace:         spec.Namespace,
			Values:            spec.Values,
			Upstreams:         upstreams,
			hasDependencies:   hasDependencies,
			KubernetesCluster: config.Resources.KubernetesCluster,
			Secrets:           helmSecretValues,
			hasChanged:        cd.HasChanged(),
		}
	}

	return artifacts, applications, config.Name, nil
}

func parseSecrets(secretConfigs []SecretConfig, secretProviderConfigs map[string]SecretProvider) ([]HelmSecretValue, error) {

	helmSecretValues := make([]HelmSecretValue, len(secretConfigs))
	for i, secretConfig := range secretConfigs {
		provider, ok := secretProviderConfigs[secretConfig.Provider]

		if !ok {
			return helmSecretValues, errors.New("Invalid secret provider reference")
		}

		helmSecretValue := HelmSecretValue{
			HelmKey:    secretConfig.HelmKey,
			SecretName: secretConfig.SecretName,
			Provider:   provider,
		}

		helmSecretValues[i] = helmSecretValue
	}
	return helmSecretValues, nil
}
