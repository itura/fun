package build

import (
	"fmt"
	"strings"
)

func (a Application) ToGitHubActionsJob(cmd string, configPath string) GitHubActionsJob {
	var needs []string

	for _, job := range a.Upstreams {
		needs = append(needs, job.JobId())
	}

	return GitHubActionsJob{
		Name:   "Deploy " + a.Id,
		RunsOn: "ubuntu-latest",
		Permissions: map[string]string{
			"id-token": "write",
			"contents": "read",
		},
		Needs: needs,
		Steps: a.GetSteps(cmd, configPath),
	}
}

func (a Application) GetSteps(cmd string, configPath string) []GitHubActionsStep {
	checkoutStep := GitHubActionsStep{
		Name: "Checkout Repo",
		Uses: "actions/checkout@v3",
		With: map[string]interface{}{
			"fetch-depth": 2,
		},
	}

	setupGoStep := GitHubActionsStep{
		Name: "Setup Go",
		Uses: "actions/setup-go@v3",
		With: map[string]interface{}{
			"go-version": "1.19",
		},
	}

	cloudProviderAuthStep := a.CloudProvider.Impl().AuthStep()

	steps := []GitHubActionsStep{
		checkoutStep,
		setupGoStep,
		cloudProviderAuthStep,
	}

	if a.Type == typeHelm {
		steps = append(steps, GetHelmSteps(a.KubernetesCluster)...)
		if len(a.Secrets) > 0 {
			secrets := Secrets{Secrets: a.Secrets, SecretProviders: a.SecretProviders}
			steps = append(steps, secrets.ToGitHubActionsSteps()...)
		}
	} else if a.Type == typeTerraform {
		steps = append(steps, GetTerraformSteps()...)
	}

	deployStep := GetDeployStep(a.Id, a.Values, a.ResolveSecrets(), GetDeployRunCommand(a.Id, cmd, configPath))

	steps = append(steps, deployStep)

	return steps
}

func GetHelmSteps(cluster ClusterConfig) []GitHubActionsStep {
	gkeAuthStep := GitHubActionsStep{
		Name: "Authenticate to GKE Cluster",
		Uses: "google-github-actions/get-gke-credentials@v1",
		With: map[string]interface{}{
			"cluster_name": cluster.Name,
			"location":     cluster.Location,
		},
	}

	setupHelmStep := GitHubActionsStep{
		Name: "Setup Helm",
		Uses: "azure/setup-helm@v3",
		With: map[string]interface{}{
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
			With: map[string]interface{}{
				"terraform_version": "1.3.6",
			},
		},
	}
}

func GetDeployStep(applicationId string, values []HelmValue, resolvedSecrets map[string]string, runCommand string) GitHubActionsStep {
	var envMap map[string]string
	if len(resolvedSecrets) > 0 || len(values) > 0 {
		envMap = map[string]string{}
		for _, helmValue := range values {
			envVarName := resolveKey(helmValue)
			envMap[envVarName] = helmValue.Value
		}
		for envVarName, secret := range resolvedSecrets {
			envMap[envVarName] = secret
		}
	}

	return GitHubActionsStep{
		Name: "Deploy " + applicationId,
		Env:  envMap,
		Run:  runCommand,
	}
}

func GetDeployRunCommand(applicationId string, cmd string, configPath string) string {
	return strings.Join([]string{
		fmt.Sprintf("go run %s deploy-application %s", cmd, applicationId),
		"--config " + configPath,
		"--current-sha $GITHUB_SHA",
	}, " \\\n  ")
}
