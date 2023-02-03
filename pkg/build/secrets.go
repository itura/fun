package build

import (
	"fmt"
	"sort"
	"strings"
)

type SecretProvider interface {
	Add(key string, secretName string) SecretProvider
	GetRuntimeArgs() []RuntimeArg
	GenerateFetchSteps() []GitHubActionsStep
	GenerateEnvMap() map[string]string
}

type GitHubActionsSecretProvider struct {
	secrets []RuntimeArg
}

func (g GitHubActionsSecretProvider) GetRuntimeArgs() []RuntimeArg {
	sort.Slice(g.secrets, func(i, j int) bool {
		return g.secrets[i].Key < g.secrets[j].Key
	})
	return g.secrets
}

func (g GitHubActionsSecretProvider) Add(key, secretName string) SecretProvider {
	g.secrets = append(g.secrets, RuntimeArg{key, g.resolve(secretName)})
	return g
}

func (g GitHubActionsSecretProvider) GenerateFetchSteps() []GitHubActionsStep {
	return []GitHubActionsStep{}
}

func (g GitHubActionsSecretProvider) GenerateEnvMap() map[string]string {
	envMap := map[string]string{}
	for _, secret := range g.secrets {
		envMap[secret.EnvKey()] = secret.Value
	}
	return envMap
}

func (g GitHubActionsSecretProvider) resolve(secretName string) string {
	return fmt.Sprintf("${{ secrets.%s }}", secretName)
}

type GcpSecretProvider struct {
	project string
	id      string
	secrets []RuntimeArg
}

func (g GcpSecretProvider) GetRuntimeArgs() []RuntimeArg {
	var results []RuntimeArg
	for _, secret := range g.secrets {
		results = append(results, RuntimeArg{secret.Key, g.resolve(secret.Value)})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Key < results[j].Key
	})
	return results
}

func (g GcpSecretProvider) Add(key, secretName string) SecretProvider {
	g.secrets = append(g.secrets, RuntimeArg{key, secretName})
	return g
}

func (g GcpSecretProvider) GenerateFetchSteps() []GitHubActionsStep {
	if len(g.secrets) == 0 {
		return []GitHubActionsStep{}
	}
	var names []string
	for _, secret := range g.secrets {
		names = append(names, secret.Value)
	}
	return []GitHubActionsStep{FetchGcpSecretsStep(g.id, g.project, names...)}
}

func (g GcpSecretProvider) GenerateEnvMap() map[string]string {
	envMap := map[string]string{}
	for _, secret := range g.secrets {
		envMap[secret.EnvKey()] = g.resolve(secret.Value)
	}
	return envMap
}

func (g GcpSecretProvider) resolve(secretName string) string {
	return fmt.Sprintf("${{ steps.secrets-%s.outputs.%s }}", g.id, secretName)
}

type Secrets struct {
	Secrets         map[string][]HelmSecretValue
	SecretProviders map[string]SecretProviderRaw
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

func (b Secrets) ToFetchSecretsGitHubActionsStep() []GitHubActionsStep {
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
