package build

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type GitHubActionsWorkflow struct {
	Name string
	On   map[string]GitHubActionsTriggerEvent // v cool https://stackoverflow.com/questions/70849190/golang-how-to-avoid-double-quoted-on-key-on-when-marshaling-struct-to-yaml
	Jobs map[string]GitHubActionsJob
}

func (g GitHubActionsWorkflow) WriteYaml(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(g)
}

type GitHubActionsJob struct {
	Name        string
	RunsOn      string `yaml:"runs-on"`
	Permissions map[string]string
	Needs       []string `yaml:"needs,omitempty"`
	Steps       []GitHubActionsStep
}

type GitHubActionsStep struct {
	Id   string                 `yaml:"id,omitempty"`
	Name string                 `yaml:"name,omitempty"`
	Uses string                 `yaml:"uses,omitempty"`
	With map[string]interface{} `yaml:"with,omitempty"`
	Env  map[string]string      `yaml:"env,omitempty"`
	Run  string                 `yaml:"run,omitempty"`
}

type GitHubActionsTriggerEvent struct {
	Branches []string
}

func CheckoutRepoStep() GitHubActionsStep {
	return GitHubActionsStep{
		Name: "Checkout Repo",
		Uses: "actions/checkout@v3",
		With: map[string]interface{}{
			"fetch-depth": 2,
		},
	}
}

func SetupGoStep() GitHubActionsStep {
	return GitHubActionsStep{
		Name: "Setup Go",
		Uses: "actions/setup-go@v3",
		With: map[string]interface{}{
			"go-version": "1.19",
		},
	}
}

func GcpAuthStep(workloadIdentityProvider, serviceAccount string) GitHubActionsStep {
	return GitHubActionsStep{
		Name: "Authenticate to GCloud via Service Account",
		Uses: "google-github-actions/auth@v1",
		With: map[string]interface{}{
			"workload_identity_provider": workloadIdentityProvider,
			"service_account":            serviceAccount,
		},
	}
}

func FetchGcpSecretsStep(id string, project string, secretNames ...string) GitHubActionsStep {
	var formattedSecretNames []string
	for _, secretName := range secretNames {
		formattedSecretNames = append(formattedSecretNames, fmt.Sprintf(
			"%s:%s/%s",
			secretName,
			project,
			secretName),
		)
	}
	return GitHubActionsStep{
		Name: fmt.Sprintf("Get Secrets from GCP Provider %s", id),
		Id:   "secrets-" + id,
		Uses: "google-github-actions/get-secretmanager-secrets@v1",
		With: map[string]interface{}{
			"secrets": strings.Join(formattedSecretNames, "\n"),
		},
	}
}
