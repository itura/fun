package build

import (
	"fmt"
	"github.com/itura/fun/pkg/fun"
	"strings"
)

type RuntimeArg struct {
	Key   string
	Value string
}

func (r RuntimeArg) EnvKey() string {
	return strings.ReplaceAll(r.Key, ".", "_")
}

type Application struct {
	Id                string
	Path              string
	Repository        string
	KubernetesCluster ClusterConfig
	CurrentSha        string
	Namespace         string
	RuntimeArgs       []RuntimeArg
	Type              ApplicationType
	hasChanged        bool
	Steps             []GitHubActionsStep
}

func CreateApplications(
	args ActionArgs,
	cd ChangeDetection,
	config PipelineConfigRaw,
	artifactRepository string,
	dependencies Dependencies,
) (map[string]Application, error) {
	applications := make(map[string]Application)
	for _, spec := range config.Applications {
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
		var secretProviders = config.Resources.SecretProviders
		for entry := range fun.MapEntriesOrdered(secretProviders) {
			id, secretProviderConfig := entry.Get()
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

		applications[spec.Id] = Application{
			Type:              spec.Type,
			Id:                spec.Id,
			Path:              spec.Path,
			Repository:        artifactRepository,
			CurrentSha:        args.CurrentSha,
			Namespace:         spec.Namespace,
			RuntimeArgs:       runTimeArgs,
			KubernetesCluster: config.Resources.KubernetesCluster,
			hasChanged:        cd.HasChanged(dependencies.GetAllPaths(spec.Id)...),
			Steps:             setupSteps,
		}
	}
	return applications, nil
}

func (a Application) PrepareBuild() Build {
	switch a.Type {
	case applicationTypeTerraform:
		return NewTerraform(a)
	case applicationTypeHelm:
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
