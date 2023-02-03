package build

import (
	"fmt"
	"sort"
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
