package build

type SecretProvider1 interface {
	Add(key string, secretName string) SecretProvider
	GetValue(id string) string
	GetSetupSteps() []GitHubActionsStep
}

type CloudProvider1[StepDto any] interface {
	GetSetupSteps() []StepDto
	Validate(key string) ValidationErrors
}

type ArtifactRepositoryProvider[StepDto any] interface {
	GetSetupSteps() []StepDto
	Validate(key string) ValidationErrors
}

type ProviderFactory[StepDto any] interface {
	GetCloudProvider(t CloudProviderType) CloudProvider1[StepDto]
	GetArtifactRepositoryProvider(t string) ArtifactRepositoryProvider[StepDto]
	GetSecretProviders(t SecretProviderType) []SecretProvider1
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
	//	factory := GithubActionsProviderFactory{config: config}
	//
	//	cloudProvider := factory.GetCloudProvider(config.Resources.CloudProvider.Type)
	//	artifactRepositoryProvider := factory.GetArtifactRepositoryProvider()
	//
	//	setupSteps := append(
	//		[]GitHubActionsStep{CheckoutRepoStep(), SetupGoStep()},
	//		cloudProvider.GetSetupSteps()...,
	//	)
	//
	//} else if buildType == "circleCi" {
	//	factory := CircleCiProviderFactory{config: config}
	//	cloudProvider := factory.GetCloudProvider(config.Resources.CloudProvider.Type)
	//
	//	blah := cloudProvider.GetSetupSteps()
	//}

}

type CircleCiStep struct{}
