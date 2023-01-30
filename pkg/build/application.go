package build

import (
	"fmt"
	"strings"
)

type HelmValue struct {
	Key   string
	Value string
}

type HelmSecretValue struct {
	HelmKey    string
	SecretName string
}

type Application struct {
	Id                string
	Path              string
	Repository        string
	KubernetesCluster ClusterConfig
	CurrentSha        string
	Namespace         string
	Values            []HelmValue
	Upstreams         []Job
	Type              ApplicationType
	Secrets           map[string][]HelmSecretValue
	SecretProviders   map[string]SecretProvider
	hasDependencies   bool
	hasChanged        bool
	CloudProvider     CloudProviderConfig
}

func CreateApplications(args ActionArgs, previousSha string, config PipelineConfig, artifacts map[string]Artifact, repository string) (map[string]Application, error) {
	applications := make(map[string]Application)
	for _, spec := range config.Applications {
		var upstreams []Job
		var cd ChangeDetection
		if args.Force {
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

		helmSecretValues := make(map[string][]HelmSecretValue, len(secretConfigs))
		for _, secretConfig := range secretConfigs {
			_, ok := config.Resources.SecretProviders[secretConfig.Provider]

			if !ok {
				return nil, MissingSecretProvider{}
			}

			helmSecretValue := HelmSecretValue{
				HelmKey:    secretConfig.HelmKey,
				SecretName: secretConfig.SecretName,
			}

			providerSecretsList := helmSecretValues[secretConfig.Provider]
			helmSecretValues[secretConfig.Provider] = append(providerSecretsList, helmSecretValue)
		}

		hasDependencies := len(spec.Dependencies) > 0 || len(spec.Artifacts) > 0
		applications[spec.Id] = Application{
			Type:              spec.Type,
			Id:                spec.Id,
			Path:              spec.Path,
			Repository:        repository,
			CurrentSha:        args.CurrentSha,
			Namespace:         spec.Namespace,
			Values:            spec.Values,
			Upstreams:         upstreams,
			hasDependencies:   hasDependencies,
			KubernetesCluster: config.Resources.KubernetesCluster,
			Secrets:           helmSecretValues,
			SecretProviders:   config.Resources.SecretProviders,
			hasChanged:        cd.HasChanged(),
			CloudProvider:     config.Resources.CloudProvider,
		}
	}
	return applications, nil
}

func (a Application) PrepareBuild() (Build, error) {
	switch a.Type {
	case typeHelm:
		return NewHelm(a), nil
	case typeTerraform:
		return NewTerraform(a), nil
	default:
		return NullBuild{}, fmt.Errorf("invalid application type %s", a.Type)
	}
}

func (a Application) JobId() string {
	return fmt.Sprintf("deploy-%s", a.Id)
}

func (a Application) HasDependencies() bool {
	return a.hasDependencies
}

func (a Application) AddValue(key, value string) Application {
	a.Values = append(a.Values, HelmValue{
		Key:   key,
		Value: value,
	})
	return a
}

func (a Application) SetSecret(key, provider, name string) Application {
	value, present := a.Secrets[provider]
	if !present {
		value = []HelmSecretValue{}
	}
	a.Secrets[provider] = append(value, HelmSecretValue{
		HelmKey:    key,
		SecretName: name,
	})
	return a
}

func (a Application) ResolveSecrets() map[string]string {
	secretMappings := map[string]string{}

	for providerId, secrets := range a.Secrets {
		provider := a.SecretProviders[providerId]
		for _, secret := range secrets {
			envVarName := strings.ReplaceAll(secret.HelmKey, ".", "_")
			switch provider.Type {
			case secretProviderTypeGcp:
				secretValue := fmt.Sprintf("${{ steps.secrets-%s.outputs.%s }}", providerId, secret.SecretName)

				secretMappings[envVarName] = secretValue
			case secretProviderTypeGithub:
				secretValue := fmt.Sprintf("${{ secrets.%s }}", secret.SecretName)

				secretMappings[envVarName] = secretValue
			}
		}
	}
	return secretMappings
}
