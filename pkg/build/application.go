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

	secretProviders := NewSecretProviders1(config.Resources.SecretProviders)
	applications := make(map[string]Application)
	allValidationErrors := NewValidationErrors("applications")

	for _, spec := range config.Applications {

		applicationConfigErrors := NewValidationErrors(spec.Id)
		applicationConfigErrors = secretProviders.Validate(applicationConfigErrors, spec.Secrets)
		if applicationConfigErrors.IsPresent() {
			allValidationErrors = allValidationErrors.PutChild(applicationConfigErrors)
			continue
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

		runTimeArgs = append(runTimeArgs, secretProviders.ResolveRuntimeArgs(spec.Secrets)...)
		setupSteps = append(setupSteps, secretProviders.ResolveSetupSteps(spec.Secrets)...)

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

	if allValidationErrors.IsPresent() {
		return map[string]Application{}, allValidationErrors
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
