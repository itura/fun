package build

import (
	"fmt"
	"strings"
)

type RuntimeArg struct {
	Key   string
	Value string
}

func (r RuntimeArg) EnvKey() string {
	return strings.ReplaceAll(r.Key, ".", "_")
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
	RuntimeArgs       []RuntimeArg
	Upstreams         []Job
	Type              ApplicationType
	Secrets           map[string][]HelmSecretValue
	SecretProviders   map[string]SecretProviderRaw
	hasDependencies   bool
	hasChanged        bool
	CloudProvider     CloudProviderConfig
	Steps             []GitHubActionsStep
}

func CreateApplications(args ActionArgs, previousSha string, config PipelineConfigRaw, artifacts map[string]Artifact, artifactRepository string) (map[string]Application, error) {
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

		var runTimeArgs []RuntimeArg
		for _, arg := range spec.Values {
			runTimeArgs = append(runTimeArgs, RuntimeArg{
				Key:   arg.Key,
				Value: arg.Value,
			})
		}
		setupSteps := []GitHubActionsStep{
			CheckoutRepoStep(),
			SetupGoStep(),
			config.Resources.CloudProvider.Impl().AuthStep(),
		}

		numFound := 0
		for id, secretProviderConfig := range config.Resources.SecretProviders {
			secretProvider := secretProviderConfig.Impl(id)
			for _, secretConfig := range spec.Secrets {
				if secretConfig.Provider == id {
					numFound++
					secretProvider = secretProvider.Add(secretConfig.HelmKey, secretConfig.SecretName)
				}
			}
			runTimeArgs = append(runTimeArgs, secretProvider.GetRuntimeArgs()...)
			setupSteps = append(setupSteps, secretProvider.GenerateFetchSteps()...)
		}

		if numFound != len(spec.Secrets) {
			return nil, MissingSecretProvider{}
		}

		hasDependencies := len(spec.Dependencies) > 0 || len(spec.Artifacts) > 0
		applications[spec.Id] = Application{
			Type:              spec.Type,
			Id:                spec.Id,
			Path:              spec.Path,
			Repository:        artifactRepository,
			CurrentSha:        args.CurrentSha,
			Namespace:         spec.Namespace,
			RuntimeArgs:       runTimeArgs,
			Upstreams:         upstreams,
			hasDependencies:   hasDependencies,
			KubernetesCluster: config.Resources.KubernetesCluster,
			hasChanged:        cd.HasChanged(),
			CloudProvider:     config.Resources.CloudProvider,
			Steps:             setupSteps,
		}
	}
	return applications, nil
}

func (a Application) PrepareBuild() Build {
	switch a.Type {
	case typeTerraform:
		return NewTerraform(a)
	case typeHelm:
		return NewHelm(a)
	default:
		return NullBuild{}
	}
}

func (a Application) JobId() string {
	return fmt.Sprintf("deploy-%s", a.Id)
}

func (a Application) AddRuntimeArg(key, value string) Application {
	a.RuntimeArgs = append(a.RuntimeArgs, RuntimeArg{
		Key:   key,
		Value: value,
	})
	return a
}

func (a Application) SetNamespace(namespace string) Application {
	a.Namespace = namespace
	return a
}

func (a Application) AddStep(steps ...GitHubActionsStep) Application {
	a.Steps = append(a.Steps, steps...)
	return a
}
