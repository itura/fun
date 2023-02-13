package build

import (
	"fmt"
	"strings"
)

func (a Application) ToGitHubActionsJob(cmd string, configPath string, dependencies Dependencies) GitHubActionsJob {
	return GitHubActionsJob{
		Name:   "Deploy " + a.Id,
		RunsOn: "ubuntu-latest",
		Permissions: map[string]string{
			"id-token": "write",
			"contents": "read",
		},
		Needs: dependencies.GetUpstreamJobIds(a.Id),
		Steps: a.GetSteps(cmd, configPath),
	}
}

func (a Application) GetSteps(cmd string, configPath string) []GitHubActionsStep {
	if a.Type == applicationTypeHelm {
		a.Steps = append(a.Steps, GetHelmSteps(a.KubernetesCluster)...)
	} else if a.Type == applicationTypeTerraform {
		a.Steps = append(a.Steps, GetTerraformSteps()...)
	}

	deployStep := GetDeployStep(a.Id, a.RuntimeArgs, GetDeployRunCommand(a.Id, cmd, configPath))

	a.Steps = append(a.Steps, deployStep)

	return a.Steps
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

func GetDeployStep(applicationId string, runtimeArgs []RuntimeArg, runCommand string) GitHubActionsStep {
	var envMap map[string]string
	if len(runtimeArgs) > 0 {
		envMap = map[string]string{}
		for _, arg := range runtimeArgs {
			envMap[arg.EnvKey()] = arg.Value
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
