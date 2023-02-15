package build

import (
	"fmt"
	"github.com/itura/fun/pkg/fun"
)

type SecretProviders1 struct {
	secretProviders []SecretProvider
}

func NewSecretProviders1(secretProviderConfigs SecretProviderConfigs) SecretProviders1 {
	var secretProviders []SecretProvider

	for _, secretProviderConfig := range secretProviderConfigs {
		secretProviders = append(secretProviders, secretProviderConfig.Impl())

	}

	return SecretProviders1{
		secretProviders: secretProviders,
	}
}

func (s SecretProviders1) Validate(validationErrors ValidationErrors, secretConfigs []SecretConfig) ValidationErrors {

	providerSecretNames := map[string]any{}
	for _, provider := range s.secretProviders {
		for _, secretName := range provider.GetSecretNames() {
			providerSecretNames[secretName] = nil
		}
	}

	for _, secretConfig := range secretConfigs {
		_, ok := providerSecretNames[secretConfig.SecretName]
		if !ok {
			validationErrors = validationErrors.PutChild(
				NewValidationErrors("secrets").Put(
					secretConfig.Key,
					fmt.Errorf("secret '%s' not configured in any secretProvider", secretConfig.SecretName),
				),
			)
		}
	}

	return validationErrors
}

func (s SecretProviders1) ResolveRuntimeArgs(secretConfigs []SecretConfig) []RuntimeArg {
	var runtimeArgs []RuntimeArg
	for _, provider := range s.secretProviders {
		runtimeArgs = append(runtimeArgs, provider.ResolveRuntimeArgs(secretConfigs)...)
	}
	return runtimeArgs
}

func (s SecretProviders1) ResolveSetupSteps(secretConfigs []SecretConfig) []GitHubActionsStep {
	var steps []GitHubActionsStep
	for _, provider := range s.secretProviders {
		steps = append(steps, provider.ResolveSetupSteps(secretConfigs)...)
	}
	return steps
}

type SecretProvider interface {
	ResolveRuntimeArgs(secretConfigs []SecretConfig) []RuntimeArg
	ResolveSetupSteps(secretConfigs []SecretConfig) []GitHubActionsStep
	GetSecretNames() []string
	Validate(ValidationErrors) ValidationErrors
}

type GitHubActionsSecretProvider struct {
	secretNames []string
}

func (g GitHubActionsSecretProvider) Validate(e ValidationErrors) ValidationErrors {
	return e
}

func (g GitHubActionsSecretProvider) GetSecretNames() []string {
	var secretNames []string
	for _, secret := range g.secretNames {
		secretNames = append(secretNames, secret)
	}
	return secretNames
}

func (g GitHubActionsSecretProvider) ResolveRuntimeArgs(secretConfigs []SecretConfig) []RuntimeArg {
	var runtimeArgs []RuntimeArg
	for _, secretConfig := range secretConfigs {
		if fun.Contains(g.secretNames, secretConfig.SecretName) {
			runtimeArgs = append(runtimeArgs, RuntimeArg{
				Key:   secretConfig.Key,
				Value: g.resolve(secretConfig.SecretName),
			})
		}
	}
	return runtimeArgs
}

func (g GitHubActionsSecretProvider) ResolveSetupSteps(secretConfigs []SecretConfig) []GitHubActionsStep {
	return []GitHubActionsStep{}
}

func (g GitHubActionsSecretProvider) resolve(secretName string) string {
	return fmt.Sprintf("${{ secrets.%s }}", secretName)
}

type GcpSecretProvider struct {
	config      map[string]string
	id          string
	secretNames []string
}

func (g GcpSecretProvider) Validate(parent ValidationErrors) ValidationErrors {
	if len(g.config) == 0 {
		return parent.Put("config", eMissingRequiredField)
	}
	if _, ok := g.config["project"]; !ok {
		return parent.PutChild(NewValidationErrors("config").
			Put("project", eMissingRequiredField),
		)
	}
	return parent
}

func (g GcpSecretProvider) GetSecretNames() []string {
	return g.secretNames
}

func (g GcpSecretProvider) ResolveRuntimeArgs(secretConfigs []SecretConfig) []RuntimeArg {
	var runtimeArgs []RuntimeArg
	for _, secretConfig := range secretConfigs {
		if fun.Contains(g.secretNames, secretConfig.SecretName) {
			runtimeArgs = append(runtimeArgs, RuntimeArg{
				Key:   secretConfig.Key,
				Value: g.resolve(secretConfig.SecretName),
			})
		}
	}
	return runtimeArgs
}

func (g GcpSecretProvider) ResolveSetupSteps(secretConfigs []SecretConfig) []GitHubActionsStep {
	if len(g.secretNames) == 0 || len(secretConfigs) == 0 {
		return []GitHubActionsStep{}
	}
	var names []string
	for _, secretConfig := range secretConfigs {
		if fun.Contains(g.secretNames, secretConfig.SecretName) {
			names = append(names, secretConfig.SecretName)
		}
	}
	return []GitHubActionsStep{FetchGcpSecretsStep(g.id, g.project(), names...)}
}

func (g GcpSecretProvider) resolve(secretName string) string {
	return fmt.Sprintf("${{ steps.secrets-%s.outputs.%s }}", g.id, secretName)
}

func (g GcpSecretProvider) project() string {
	return g.config["project"]
}
