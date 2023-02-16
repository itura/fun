package build

import (
	"fmt"
	"github.com/itura/fun/pkg/fun"
)

func validateConfig(config PipelineConfigRaw) ValidationErrors {
	validationErrors := config.Validate("")
	if validationErrors.IsPresent() {
		return validationErrors
	}

	secretProviders := NewSecretProviders1(config.Resources.SecretProviders)
	allValidationErrors := NewValidationErrors("applications")

	for _, spec := range config.Applications {

		applicationConfigErrors := NewValidationErrors(spec.Id)
		applicationConfigErrors = secretProviders.Validate(applicationConfigErrors, spec.Secrets)
		if applicationConfigErrors.IsPresent() {
			allValidationErrors = allValidationErrors.PutChild(applicationConfigErrors)
			continue
		}

	}
	return allValidationErrors
}

func ParseConfigForGeneration(configPath string, cmd string) (WorkflowWriter, error) {
	config, err := readFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("shucks")
	}

	validationErrors := validateConfig(config)
	if validationErrors.IsPresent() {
		return nil, err
	}

	buildType := "github"
	dependencies := ParseDependencies(config)

	if buildType == "github" {
		return NewGithubActionsFactory(configPath, cmd, config, dependencies).Create(), nil
	}

	return nil, nil
}

func NewGithubActionsFactory(configPath string, cmd string, config PipelineConfigRaw, dependencies Dependencies) GithubActionsFactory {
	return GithubActionsFactory{
		config:       config,
		configPath:   configPath,
		cmd:          cmd,
		dependencies: dependencies,
	}
}

type GithubActionsFactory struct {
	config       PipelineConfigRaw
	configPath   string
	cmd          string
	dependencies Dependencies
}

func (g GithubActionsFactory) Create() GitHubActionsWorkflow {
	commonSetupSteps := []GitHubActionsStep{}
	commonSetupSteps = append(
		commonSetupSteps,
		CheckoutRepoStep(),
		SetupGoStep(),
	)
	commonSetupSteps = append(
		commonSetupSteps,
		g.GetCloudProviderSteps()...,
	)

	jobs := map[string]GitHubActionsJob{}
	//for _, artifact := range g.config.Artifacts {
	//	applicationSteps := make([]GitHubActionsStep, len(commonSetupSteps))
	//	copy(applicationSteps, commonSetupSteps)
	//}
	for _, application := range g.config.Applications {
		applicationSteps := make([]GitHubActionsStep, len(commonSetupSteps))
		copy(applicationSteps, commonSetupSteps)

		applicationSteps = append(
			applicationSteps,
			g.GetDeploySteps(application, g.configPath, g.cmd)...,
		)

		jobId := g.dependencies.GetJobId(application.Id)
		jobs[jobId] = GetGitHubActionsJob(application.Id, applicationSteps, g.dependencies)
	}
	return GitHubActionsWorkflow{
		Name: g.config.Name,
		On: map[string]GitHubActionsTriggerEvent{
			"push": {
				Branches: []string{"trunk"},
			},
		},
		Jobs: jobs,
	}

}

func (g GithubActionsFactory) GetCloudProviderSteps() []GitHubActionsStep {
	providerConfig := g.config.Resources.CloudProvider
	switch providerConfig.Type {
	case cloudProviderTypeGcp:
		return []GitHubActionsStep{GcpAuthStep(
			formatSecretValue(providerConfig.Config["workloadIdentityProvider"]),
			formatSecretValue(providerConfig.Config["serviceAccount"]),
		)}
	default:
		panic("😎")
	}
}

func (g GithubActionsFactory) GetDeploySteps(application ApplicationConfig, configPath string, cmd string) []GitHubActionsStep {

	var steps []GitHubActionsStep
	var runTimeArgs []RuntimeArg
	for _, arg := range application.Values {
		runTimeArgs = append(runTimeArgs, RuntimeArg{
			Key:   arg.Key,
			Value: arg.Value,
		})
	}

	if len(application.Secrets) != 0 {

		for _, secretProviderConfig := range g.config.Resources.SecretProviders {
			secretNames := secretProviderConfig.SecretNames
			id := secretProviderConfig.Id
			project := secretProviderConfig.Config["project"]

			if len(secretNames) == 0 {
				continue
			}

			switch secretProviderConfig.Type {

			case secretProviderTypeGcp:
				var names []string
				for _, secretConfig := range application.Secrets {
					if fun.Contains(secretNames, secretConfig.SecretName) {
						names = append(names, secretConfig.SecretName)
						runTimeArgs = append(runTimeArgs, RuntimeArg{
							Key:   secretConfig.Key,
							Value: g.resolveStepOutput(fmt.Sprintf("secrets-%s", id), secretConfig.SecretName),
						})
					}
				}

				steps = append(steps, FetchGcpSecretsStep(id, project, names...))
			case secretProviderTypeGithub:
				for _, secretConfig := range application.Secrets {
					if fun.Contains(secretNames, secretConfig.SecretName) {
						runTimeArgs = append(runTimeArgs, RuntimeArg{
							Key:   secretConfig.Key,
							Value: g.resolveSecret(secretConfig.SecretName),
						})
					}
				}
				continue

			default:
				panic("👺")
			}
		}
	}

	switch application.Type {
	case applicationTypeHelm:
		switch g.config.Resources.KubernetesCluster.Type {
		case "gke":
			steps = append(steps, GetSetupGkeStep(g.config.Resources.KubernetesCluster))
		default:
			panic("🍕")
		}

		steps = append(steps, GetSetupHelmStep())
	case applicationTypeTerraform:
		steps = append(steps, GetSetupTerraformStep())
	default:
		panic("😅")
	}

	steps = append(steps, GetDeployStep(application.Id, runTimeArgs, GetDeployRunCommand(application.Id, configPath, cmd)))

	return steps
}

func (g GithubActionsFactory) GetArtifactRepositoryProvider(t string) ArtifactRepositoryProvider[GitHubActionsStep] {
	//TODO implement me
	panic("implement me")
}

func (g GithubActionsFactory) GetKubernetesProvider(t string) KubernetesProvider[GitHubActionsStep] {
	//TODO implement me
	panic("implement me")
}

func (g GithubActionsFactory) resolveStepOutput(stepId string, outputId string) string {
	return fmt.Sprintf("${{ steps.%s.outputs.%s }}", stepId, outputId)
}

func (g GithubActionsFactory) resolveSecret(secretName string) string {
	return fmt.Sprintf("${{ secrets.%s }}", secretName)
}
