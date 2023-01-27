package build

import (
	"fmt"
	"strings"
)

func (a Artifact) ToGitHubActionsJob(cmd string, configPath string) GitHubActionsJob {
	var upstreamIds []string = nil

	for _, job := range a.Upstreams {
		upstreamIds = append(upstreamIds, job.JobId())
	}

	return GitHubActionsJob{
		Name:   "Build " + a.Id,
		RunsOn: "ubuntu-latest",
		Permissions: map[string]string{
			"id-token": "write",
			"contents": "read",
		},
		Needs: upstreamIds,
		Steps: a.GetSteps(cmd, configPath),
	}
}

func (a Artifact) GetSteps(cmd string, configPath string) []GitHubActionsStep {
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

	cloudProviderAuthStep := a.CloudProvider.GetGitHubActionsAuthStep()

	setupGcloudStep := GitHubActionsStep{
		Name: "Configure GCloud SDK",
		Uses: "google-github-actions/setup-gcloud@v0",
	}

	configureDockerStep := GitHubActionsStep{
		Name: "Configure Docker",
		Run:  fmt.Sprintf("gcloud --quiet auth configure-docker %s", a.Host),
	}

	buildArtifactCommand := strings.Join(
		[]string{
			fmt.Sprintf("go run %s build-artifact %s", cmd, a.Id),
			fmt.Sprintf("--config %s", configPath),
			"--current-sha $GITHUB_SHA",
		}, " \\\n  ",
	)

	buildArtifactStep := GitHubActionsStep{
		Name: fmt.Sprintf("Build %s", a.Id),
		Run:  buildArtifactCommand,
	}

	return []GitHubActionsStep{
		checkoutStep,
		setupGoStep,
		cloudProviderAuthStep,
		setupGcloudStep,
		configureDockerStep,
		buildArtifactStep,
	}
}

// TODO: this belongs somewhere else
func (c CloudProvider) GetGitHubActionsAuthStep() GitHubActionsStep {
	if c.Type == "gcp" {
		serviceAccount := c.Config["serviceAccount"]

		workloadIdentityProvider := c.Config["workloadIdentityProvider"]

		googleAuthStep := GitHubActionsStep{
			Name: "Authenticate to GCloud via Service Account",
			Uses: "google-github-actions/auth@v1",
			With: map[string]interface{}{
				"workload_identity_provider": formatSecretValue(workloadIdentityProvider),
				"service_account":            formatSecretValue(serviceAccount),
			},
		}

		return googleAuthStep
	} else {
		return GitHubActionsStep{}
	}
}
