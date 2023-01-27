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

	googleAuthStep := GitHubActionsStep{
		Name: "Authenticate to GCloud via Service Account",
		Uses: "google-github-actions/auth@v1",
		With: map[string]interface{}{
			"workload_identity_provider": "TODO",
			"service_account":            "TODO",
		},
	}

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
			"--project-id $PROJECT_ID",
		}, " \\\n  ",
	)

	buildArtifactStep := GitHubActionsStep{
		Name: fmt.Sprintf("Build %s", a.Id),
		Env: map[string]string{
			"PROJECT_ID": a.Project,
		},
		Run: buildArtifactCommand,
	}

	return []GitHubActionsStep{
		checkoutStep,
		setupGoStep,
		googleAuthStep,
		setupGcloudStep,
		configureDockerStep,
		buildArtifactStep,
	}
}
