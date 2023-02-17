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

func ReadConfigForGeneration(configPath string, cmd string) (WorkflowWriter, error) {
	config, err := readFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("shucks")
	}

	return ParseConfigForGeneration(config, configPath, cmd)
}

func ParseConfigForGeneration(config PipelineConfigRaw, configPath string, cmd string) (WorkflowWriter, error) {
	validationErrors := validateConfig(config)
	if validationErrors.IsPresent() {
		return nil, validationErrors
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

func (g GithubActionsFactory) getCommonSetupSteps() []GitHubActionsStep {
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
	return commonSetupSteps
}

func (g GithubActionsFactory) Create() GitHubActionsWorkflow {
	workflow := NewGitHubActionsWorkflow(g.config.Name)
	for _, artifact := range g.config.Artifacts {
		jobId := g.dependencies.GetJobId(artifact.Id)
		workflow = workflow.SetJob(jobId, g.GetArtifactJob(artifact))
	}

	for _, application := range g.config.Applications {
		jobId := g.dependencies.GetJobId(application.Id)
		workflow = workflow.SetJob(jobId, g.GetApplicationJob(application))
	}
	return workflow
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

func (g GithubActionsFactory) GetArtifactJob(artifact ArtifactConfig) GitHubActionsJob {
	steps := g.getCommonSetupSteps()

	repoConfig := g.config.Resources.ArtifactRepository
	switch repoConfig.Type {
	case artifactRepositoryTypeGcpDocker:
		steps = append(steps,
			ConfigureGcloudCliStep(),
			ConfigureGcloudDockerStep(repoConfig.Host),
			BuildArtifactStep(artifact.Id, g.configPath, g.cmd),
		)
	default:
		t, _ := ArtifactRepositoryTypeEnum.ToString(repoConfig.Type)
		panic(t + "is not a valid type")
	}

	return NewGitHubActionsJob("Build " + artifact.Id).
		AddNeeds(g.dependencies.GetUpstreamJobIds(artifact.Id)...).
		AddSteps(steps...)
}

func (g GithubActionsFactory) GetApplicationJob(application ApplicationConfig) GitHubActionsJob {
	steps := g.getCommonSetupSteps()

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

	steps = append(steps, GetDeployStep(application.Id, runTimeArgs, GetDeployRunCommand(application.Id, g.cmd, g.configPath)))

	return NewGitHubActionsJob("Deploy " + application.Id).
		AddNeeds(g.dependencies.GetUpstreamJobIds(application.Id)...).
		AddSteps(steps...)
}

func (g GithubActionsFactory) resolveStepOutput(stepId string, outputId string) string {
	return fmt.Sprintf("${{ steps.%s.outputs.%s }}", stepId, outputId)
}

func (g GithubActionsFactory) resolveSecret(secretName string) string {
	return fmt.Sprintf("${{ secrets.%s }}", secretName)
}
