package build

import (
	"fmt"
	"strings"
)

func (a Application) ToGitHubActionsJob() GitHubActionsJob {
	return GitHubActionsJob{
		Name:   "Deploy " + a.Id,
		RunsOn: "ubuntu-latest",
		Permissions: map[string]string{
			"id-token": "write",
			"contents": "read",
		},
		Steps: a.GetSteps(),
	}
}

func (a Application) GetSteps() []GitHubActionsStep {
	checkoutStep := GitHubActionsStep{
		Name: "Checkout Repo",
		Uses: "actions/checkout@v3",
		With: map[string]string{
			"fetch-depth": "2",
		},
	}

	setupGoStep := GitHubActionsStep{
		Name: "Setup Go",
		Uses: "actions/setup-go@v3",
		With: map[string]string{
			"go-version": "1.19",
		},
	}

	googleAuthStep := GitHubActionsStep{
		Name: "Authenticate to GCloud via Service Account",
		Uses: "google-github-actions/auth@v1",
		With: map[string]string{
			"workload_identity_provider": "TODO",
			"service_account":            "TODO",
		},
	}

	steps := []GitHubActionsStep{
		checkoutStep,
		setupGoStep,
		googleAuthStep,
	}

	if a.Type == typeHelm {
		steps = append(steps, GetHelmSteps(a.KubernetesCluster)...)
		if len(a.Secrets) > 0 {
			steps = append(steps, GetGcpSecretsSteps(a.SecretProviders, a.Secrets)...)
		}
	} else if a.Type == typeTerraform {
		steps = append(steps, GetTerraformSteps()...)
	}

	return steps
}

func GetHelmSteps(cluster ClusterConfig) []GitHubActionsStep {
	gkeAuthStep := GitHubActionsStep{
		Name: "Authenticate to GKE Cluster",
		Uses: "google-github-actions/get-gke-credentials@v1",
		With: map[string]string{
			"cluster_name": cluster.Name,
			"location":     cluster.Location,
		},
	}

	setupHelmStep := GitHubActionsStep{
		Name: "Setup Helm",
		Uses: "azure/setup-helm@v3",
		With: map[string]string{
			"version": "v3.10.2",
		},
	}

	return []GitHubActionsStep{
		gkeAuthStep,
		setupHelmStep,
	}
}

func GetTerraformSteps() []GitHubActionsStep {
	return []GitHubActionsStep{
		{
			Name: "Setup Terraform",
			Uses: "hashicorp/setup-terraform@v2",
			With: map[string]string{
				"terraform_version": "1.3.6",
			},
		},
	}
}

func GetGcpSecretsSteps(providers map[string]SecretProvider, secrets map[string][]HelmSecretValue) []GitHubActionsStep {
	gcpSecretsSteps := []GitHubActionsStep{}

	for providerId, providerSecrets := range secrets {
		if len(providerSecrets) > 0 {
			provider := providers[providerId]
			if provider.Type == typeGcp {
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
					With: map[string]string{
						"secrets": strings.Join(secrets, "\n"),
					},
				}
				gcpSecretsSteps = append(gcpSecretsSteps, step)
			}
		}
	}

	return gcpSecretsSteps
}
