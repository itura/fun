package build

type StepProvider[StepDto any] interface {
	ResolveSetupSteps(secretConfigs []SecretConfig) []StepDto
	Validate(ValidationErrors) ValidationErrors
}

type SecretProvider1[StepDto any] interface {
	ResolveSetupSteps(secretConfigs []SecretConfig) []StepDto
	Validate(ValidationErrors) ValidationErrors
}

type CloudProvider1[StepDto any] interface {
	ResolveSetupSteps(secretConfigs []SecretConfig) []StepDto
	Validate(ValidationErrors) ValidationErrors
}

type ArtifactRepositoryProvider[StepDto any] interface {
	ResolveSetupSteps(secretConfigs []SecretConfig) []StepDto
	Validate(ValidationErrors) ValidationErrors
}

type KubernetesProvider[StepDto any] interface {
	ResolveSetupSteps(secretConfigs []SecretConfig) []StepDto
	Validate(ValidationErrors) ValidationErrors
}

type ProviderFactory[StepDto any] interface {
	GetCloudProvider() CloudProvider1[StepDto]
	GetSecretProvider(t SecretProviderType) SecretProvider1[StepDto]
	GetArtifactRepositoryProvider(t string) ArtifactRepositoryProvider[StepDto]
	GetKubernetesProvider(t string) KubernetesProvider[StepDto]
}

func parseConfig1(args ActionArgs, cd ChangeDetection) {
	config, err := readFile(args.ConfigPath)
	if err != nil {
		panic("😎")
	}

	validationErrors := config.Validate("")
	if validationErrors.IsPresent() {
		panic("😎")
	}
	//buildType := "github"

	//if buildType == "github" {
	//	factory := GithubActionsFactory{config: config}
	//
	//	cloudProvider := factory.GetCloudProviderSteps(config.Resources.CloudProvider.Type)
	//	artifactRepositoryProvider := factory.GetArtifactRepositoryProvider()
	//
	//	setupSteps := append(
	//		[]GitHubActionsStep{CheckoutRepoStep(), SetupGoStep()},
	//		cloudProvider.GetSetupSteps()...,
	//	)
	//
	//} else if buildType == "circleCi" {
	//	factory := CircleCiProviderFactory{config: config}
	//	cloudProvider := factory.GetCloudProviderSteps(config.Resources.CloudProvider.Type)
	//
	//	blah := cloudProvider.GetSetupSteps()
	//}

}

type CircleCiStep struct{}
