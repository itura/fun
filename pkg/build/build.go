package build

import (
	"fmt"
	"sort"
	"strings"
)

type Build interface {
	Build() (SideEffects, error)
}
type NullBuild struct{}

func (b NullBuild) Build() (SideEffects, error) {
	return SideEffects{}, fmt.Errorf("invalid build")
}

type PackageVerifier struct {
	Artifact
	dockerfile string
	workdir    string
	buildArgs  []string
}

func NewPackageVerifier(a Artifact) PackageVerifier {
	return PackageVerifier{
		Artifact:   a,
		dockerfile: fmt.Sprintf("%s/Dockerfile", a.Path),
		workdir:    a.Path,
	}
}

func (b PackageVerifier) Build() (SideEffects, error) {
	if b.hasChanged {
		return SideEffects{Commands: []Command{
			{
				Name: "docker",
				Arguments: []string{
					"build",
					"-f", b.dockerfile,
					"-t", b.VerifyImageName(),
					"--target", b.VerifyTarget(),
					b.workdir,
				},
			},
			{
				Name:      "docker",
				Arguments: []string{"run", "--rm", b.VerifyImageName()},
			},
		}}, nil

	}

	return SideEffects{}, nil
}

type DockerImage struct {
	PackageVerifier
}

func NewDockerImage(a Artifact) DockerImage {
	return DockerImage{
		PackageVerifier: NewPackageVerifier(a),
	}
}

func (b DockerImage) Build() (SideEffects, error) {
	sideEffects, err := b.PackageVerifier.Build()
	if err != nil {
		return SideEffects{}, err
	}

	commitTag := b.AppImageName(b.CurrentSha)
	greenTag := b.AppImageName(b.GreenTag())

	if b.hasChanged {
		sideEffects = sideEffects.Add(
			NewCommand("docker", "build",
				"-f", b.dockerfile,
				"-t", b.AppImageName(b.CurrentSha),
				"--target", b.AppTarget(),
				b.workdir,
			),
			NewCommand("docker", "tag", commitTag, greenTag),
			NewCommand("docker", "push",
				"--all-tags",
				b.AppImageBase(),
			),
		)
	} else {
		sideEffects = sideEffects.Add(
			NewCommand("docker", "pull", greenTag),
			NewCommand("docker", "tag", greenTag, commitTag),
			NewCommand("docker", "push", commitTag),
		)
	}
	return sideEffects, nil
}

type HelmDeployment struct {
	Application
}

func NewHelm(a Application) HelmDeployment {
	return HelmDeployment{
		Application: a,
	}
}

func (b HelmDeployment) Build() (SideEffects, error) {
	args := []string{
		"upgrade", b.Id, b.Path,
		"--install",
		"--atomic",
		"--namespace", b.Namespace,
		"--set", fmt.Sprintf("repo=%s", b.Repository),
		"--set", fmt.Sprintf("tag=%s", b.CurrentSha),
	}

	for _, value := range b.Values {
		envValue := strings.ReplaceAll(value.Key, ".", "_")
		args = append(args, "--set", fmt.Sprintf("%s=$%s", value.Key, envValue))
	}

	secrets := Secrets{Secrets: b.Secrets}

	secretArguments := secrets.SercetsToCLIArgs()
	args = append(args, secretArguments...)

	return SideEffects{
		Commands: []Command{
			{
				Name: "helm",
				Arguments: []string{
					"dep",
					"update",
				},
			},
			{
				Name:      "helm",
				Arguments: args,
			},
		},
	}, nil
}

type Secrets struct {
	Secrets         map[string][]HelmSecretValue
	SecretProviders map[string]SecretProvider
}

func (b Secrets) SercetsToCLIArgs() []string {

	secretsList := []HelmSecretValue{}

	for _, providerSecret := range b.Secrets {
		for _, secret := range providerSecret {
			secretsList = append(secretsList, secret)
		}
	}

	sort.Slice(secretsList, func(i, j int) bool {
		return secretsList[i].HelmKey < secretsList[j].HelmKey
	})

	arguments := []string{}

	for _, secret := range secretsList {
		envVarName := strings.ReplaceAll(secret.HelmKey, ".", "_")
		arguments = append(arguments, "--set", fmt.Sprintf("%s=$%s", secret.HelmKey, envVarName))
	}

	return arguments
}

func (b Secrets) Resolve() map[string]string {
	secretMappings := map[string]string{}

	for providerId, secrets := range b.Secrets {
		provider := b.SecretProviders[providerId]
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

func (b Secrets) ToGitHubActionsSteps() []GitHubActionsStep {
	gcpSecretsSteps := []GitHubActionsStep{}

	for providerId, providerSecrets := range b.Secrets {
		if len(providerSecrets) > 0 {
			provider := b.SecretProviders[providerId]
			if provider.Type == secretProviderTypeGcp {
				secrets := []string{}

				for _, secret := range providerSecrets {
					secrets = append(
						secrets,
						fmt.Sprintf("%s:%s/%s", secret.SecretName, provider.Config["project"], secret.SecretName),
					)
				}

				step := GitHubActionsStep{
					Name: fmt.Sprintf("Get Secrets from GCP Provider %s", providerId),
					Id:   "secrets-" + providerId,
					Uses: "google-github-actions/get-secretmanager-secrets@v1",
					With: map[string]interface{}{
						"secrets": strings.Join(secrets, "\n"),
					},
				}
				gcpSecretsSteps = append(gcpSecretsSteps, step)
			}
		}
	}

	return gcpSecretsSteps
}

type TfConfig struct {
	Application
}

func NewTerraform(a Application) TfConfig {
	return TfConfig{
		Application: a,
	}
}

func (b TfConfig) Build() (SideEffects, error) {
	return SideEffects{
		Commands: []Command{
			{
				Name: "terraform",
				Arguments: []string{
					"-chdir=terraform/main",
					"init",
				},
			},
			{
				Name: "terraform",
				Arguments: []string{
					"-chdir=terraform/main",
					"plan",
					"-out=plan.out",
				},
			},
			{
				Name: "terraform",
				Arguments: []string{
					"-chdir=terraform/main",
					"apply",
					"plan.out",
				},
			},
		},
	}, nil
}
