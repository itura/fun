package build

import (
	"fmt"
)

func (a Artifact) ToGitHubActionsJob() GitHubActionsJob {
	return GitHubActionsJob{
		Name:   "Build " + a.Id,
		RunsOn: "ubuntu-latest",
		Permissions: map[string]string{
			"id-token": "write",
			"contents": "read",
		},
		Steps: a.GetSteps(),
	}
}

func (a Artifact) GetSteps() []GitHubActionsStep {
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

	setupGcloudStep := GitHubActionsStep{
		Name: "Configure GCloud SDK",
		Uses: "google-github-actions/setup-gcloud@v0",
	}

	configureDockerStep := GitHubActionsStep{
		Name: "Configure Docker",
		Run:  fmt.Sprintf("gcloud --quiet auth configure-docker %s", a.Host),
	}

	// buildArtifactCommand := strings.Join(
	// 	[]string{
	// 		fmt.Sprintf("\ngo run %s build-artifact %s", "TODO-> Cmd", a.Id),
	// 		fmt.Sprintf("--config %s", "TODO-> ConfigPath"),
	// 		"--current-sha $GITHUB_SHA",
	// 		"--project-id $PROJECT_ID",
	// 	}, " \t\\\n",
	// )

	// buildArtifactStep := GitHubActionsStep{
	// 	Name: fmt.Sprintf("Build %s", a.Id),
	// 	Env: map[string]string{
	// 		"PROJECT_ID": a.Project,
	// 	},
	// 	Run: buildArtifactCommand,
	// }

	return []GitHubActionsStep{
		checkoutStep,
		setupGoStep,
		googleAuthStep,
		setupGcloudStep,
		configureDockerStep,
		// buildArtifactStep,
	}
}
