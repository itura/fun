package build

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type HelmValue struct {
	Key   string
	Value string
}

type HelmSecretValue struct {
	HelmKey    string
	SecretName string
}

type Application struct {
	Id                string
	Path              string
	ProjectId         string
	Repository        string
	KubernetesCluster ClusterConfig
	CurrentSha        string
	Namespace         string
	Values            []HelmValue
	Upstreams         []Job
	Type              ApplicationType
	Secrets           map[string][]HelmSecretValue
	SecretProviders   map[string]SecretProvider
	hasDependencies   bool
	hasChanged        bool
}

func (a Application) PrepareBuild() (Build, error) {
	switch a.Type {
	case typeHelm:
		return NewHelm(a), nil
	case typeTerraform:
		return NewTerraform(a), nil
	default:
		return NullBuild{}, fmt.Errorf("invalid application type %s", a.Type)
	}
}

func (a Application) JobId() string {
	return fmt.Sprintf("deploy-%s", a.Id)
}

func (a Application) HasDependencies() bool {
	return a.hasDependencies
}

func (a Application) Setup() string {

	if a.Type == typeHelm {

		setupSteps := fmt.Sprintf(`    - uses: google-github-actions/get-gke-credentials@v1
      with:
        cluster_name: %s
        location: %s
    - uses: azure/setup-helm@v3
      with:
        version: v3.10.2
`, a.KubernetesCluster.Name, a.KubernetesCluster.Location)

		gcpSecretsSteps, _ := GenerateGcpSecretsSteps(a.SecretProviders, a.Secrets)
		setupSteps += gcpSecretsSteps

		return setupSteps
	} else if a.Type == typeTerraform {
		return `
    - uses: 'hashicorp/setup-terraform@v2'
      with:
        terraform_version: '1.3.6'
`
	} else {
		return ""
	}
}

func (a Application) AddValue(key, value string) Application {
	a.Values = append(a.Values, HelmValue{
		Key:   key,
		Value: value,
	})
	return a
}

func (a Application) SetSecret(key, provider, name string) Application {
	value, present := a.Secrets[provider]
	if !present {
		value = []HelmSecretValue{}
	}
	a.Secrets[provider] = append(value, HelmSecretValue{
		HelmKey:    key,
		SecretName: name,
	})
	return a
}

func GenerateGcpSecretsSteps(providers map[string]SecretProvider, secrets map[string][]HelmSecretValue) (string, error) {
	if len(secrets) == 0 {
		return "", nil
	}

	gcpSecretsSteps := []GitHubActionsStep{}

	for providerId, providerSecrets := range secrets {
		if len(providerSecrets) > 0 {
			provider := providers[providerId]
			if provider.Type == typeGcp {
				secretsString := ""

				for _, secret := range providerSecrets {
					secretsString += fmt.Sprintf("\n%s:%s/%s", secret.SecretName, provider.Config["project"], secret.SecretName)
				}

				step := GitHubActionsStep{
					Id:   "secrets-" + providerId,
					Uses: "google-github-actions/get-secretmanager-secrets@v1",
					With: map[string]string{
						"secrets": secretsString,
					},
				}
				gcpSecretsSteps = append(gcpSecretsSteps, step)
			}
		}
	}

	if len(gcpSecretsSteps) > 0 {
		gcpSecretStepsStr, err := MarshalAndIndentSteps(gcpSecretsSteps)
		if err != nil {
			return "", err
		}
		return gcpSecretStepsStr, nil
	} else {
		return "", nil
	}
}

func MarshalAndIndentSteps(steps []GitHubActionsStep) (string, error) {
	stepsBytes, err := yaml.Marshal(steps)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(stepsBytes), "\n")
	for i, line := range lines {
		if i == len(lines)-1 {
			break
		}
		lines[i] = "    " + line
	}

	stepString := strings.Join(lines, "\n")
	stepString = strings.ReplaceAll(stepString, "secrets: |4-\n", "secrets: |-\n")
	return stepString, nil
}

func (a Application) ResolveSecrets() map[string]string {
	secretMappings := map[string]string{}

	for providerId, secrets := range a.Secrets {
		provider := a.SecretProviders[providerId]
		for _, secret := range secrets {
			envVarName := strings.ReplaceAll(secret.HelmKey, ".", "_")
			switch provider.Type {
			case typeGcp:
				secretValue := fmt.Sprintf("${{ steps.secrets-%s.outputs.%s }}", providerId, secret.SecretName)

				secretMappings[envVarName] = secretValue
			case typeGithub:
				secretValue := fmt.Sprintf("${{ secrets.%s }}", secret.SecretName)

				secretMappings[envVarName] = secretValue
			}
		}
	}
	return secretMappings
}
