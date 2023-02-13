package build

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowGeneration(t *testing.T) {
	builder := NewTestBuilder()
	expectedYamlBytes, _ := os.ReadFile("test_fixtures/valid_workflow.yaml")
	expectedWorkflow := GitHubActionsWorkflow{}
	err := yaml.Unmarshal(expectedYamlBytes, &expectedWorkflow)
	assert.Nil(t, err)

	piplineConfig := ValidPipelineConfig(builder)
	pipeline := NewPipeline(piplineConfig, "test_fixtures/valid_pipeline_config.yaml", "github.com/itura/fun/cmd/build@v0.1.23")

	workflow := pipeline.ToGitHubWorkflow()

	assert.Equal(t, expectedWorkflow, workflow)
}

func TestWorkflowGenerationE2e(t *testing.T) {
	expectedYamlBytes, _ := os.ReadFile("test_fixtures/valid_workflow.yaml")
	expectedWorkflow := GitHubActionsWorkflow{}
	err := yaml.Unmarshal(expectedYamlBytes, &expectedWorkflow)
	assert.Nil(t, err)

	pipeline, err := ParsePipeline(
		TestArgs("test_fixtures/valid_pipeline_config.yaml"),
		NewAlwaysChanged(),
	)
	assert.Nil(t, err)
	assert.Equal(t, expectedWorkflow, pipeline.ToGitHubWorkflow())
}

func TestDeployTerraformApplication(t *testing.T) {
	builder := NewTestBuilder()

	terraformApp := builder.Application("infra", "terraform/main", typeTerraform)
	parsedConfig := SuccessfulParse(
		"My Build",
		map[string]Artifact{},
		map[string]Application{
			"infra": terraformApp,
		}, NewDependencies(),
	)
	pipeline := NewPipeline(parsedConfig, "test_fixtures/valid_pipeline_config.yaml", "github.com/itura/fun/cmd/build@v0.1.19")

	sideEffects, err := pipeline.DeployApplication("infra")

	assert.Nil(t, err)
	assert.Equal(t, sideEffects.Commands, []Command{
		{
			Name: "terraform",
			Arguments: []string{
				"-chdir=terraform/main",
				"init",
			},
		},
		{
			Name: "terraform",
			Arguments: []string{
				"-chdir=terraform/main",
				"plan",
				"-out=plan.out",
			}},
		{
			Name: "terraform",
			Arguments: []string{
				"-chdir=terraform/main",
				"apply",
				"plan.out",
			}},
	})

}

func TestDeployHelmApplication(t *testing.T) {
	builder := NewTestBuilder()

	dbApp := PostgresHelmChart(builder)
	parsedConfig := SuccessfulParse(
		"My Build",
		map[string]Artifact{},
		map[string]Application{
			"db": dbApp,
		},
		NewDependencies(),
	)
	pipeline := NewPipeline(parsedConfig, "test_fixtures/valid_pipeline_config.yaml", "github.com/itura/fun/cmd/build@v0.1.19")

	sideEffects, err := pipeline.DeployApplication("db")

	assert.Nil(t, err)
	assert.Equal(t, []Command{
		{
			Name: "helm",
			Arguments: []string{
				"dep",
				"update",
			},
		},
		{
			Name: "helm",
			Arguments: []string{
				"upgrade",
				"db",
				"helm/db",
				"--install",
				"--atomic",
				"--namespace",
				"db-namespace",
				"--set",
				fmt.Sprintf("repo=%s", builder.repository()),
				"--set",
				"tag=currentSha",
				"--set",
				"postgresql.dbName=$postgresql_dbName",
				"--set",
				"postgresql.auth.password=$postgresql_auth_password",
				"--set",
				"postgresql.auth.username=$postgresql_auth_username",
			}},
	}, sideEffects.Commands)

}

func TestBuildChangedApplicationArtifact(t *testing.T) {
	builder := NewTestBuilder()

	clientArtifact := builder.Artifact("client", "pkgs/client")
	parsedConfig := SuccessfulParse(
		"My Build",
		map[string]Artifact{
			"client": clientArtifact,
		},
		map[string]Application{},
		NewDependencies(),
	)
	pipeline := NewPipeline(parsedConfig, "test_fixtures/valid_pipeline_config.yaml", "github.com/itura/fun/cmd/build@v0.1.19")

	sideEffects, err := pipeline.BuildArtifact("client")

	assert.Nil(t, err)
	assert.Equal(t, []Command{
		{
			Name: "docker",
			Arguments: []string{
				"build",
				"-f", fmt.Sprintf("%s/Dockerfile", clientArtifact.Path),
				"-t", "us-central1-docker.pkg.dev/gcp-project/repo-name/client-test:currentSha",
				"--target", "test",
				"pkgs/client",
			},
		},
		{
			Name: "docker",
			Arguments: []string{
				"run",
				"--rm",
				"us-central1-docker.pkg.dev/gcp-project/repo-name/client-test:currentSha",
			},
		},
		{
			Name: "docker",
			Arguments: []string{
				"build",
				"-f", fmt.Sprintf("%s/Dockerfile", clientArtifact.Path),
				"-t", "us-central1-docker.pkg.dev/gcp-project/repo-name/client-app:currentSha",
				"--target", "app",
				"pkgs/client",
			},
		},
		{
			Name: "docker",
			Arguments: []string{
				"tag",
				"us-central1-docker.pkg.dev/gcp-project/repo-name/client-app:currentSha",
				"us-central1-docker.pkg.dev/gcp-project/repo-name/client-app:latest-green",
			},
		},
		{
			Name: "docker",
			Arguments: []string{
				"push",
				"--all-tags",
				"us-central1-docker.pkg.dev/gcp-project/repo-name/client-app",
			},
		},
	}, sideEffects.Commands)
}

func TestBuildUnchangedApplicationArtifact(t *testing.T) {
	builder := NewTestBuilder()

	clientArtifact := builder.Artifact("client", "pkgs/client")
	clientArtifact.hasChanged = false
	parsedConfig := SuccessfulParse(
		"My Build",
		map[string]Artifact{
			"client": clientArtifact,
		},
		map[string]Application{},
		NewDependencies(),
	)
	pipeline := NewPipeline(parsedConfig, "test_fixtures/valid_pipeline_config.yaml", "github.com/itura/fun/cmd/build@v0.1.19")

	sideEffects, err := pipeline.BuildArtifact("client")

	assert.Nil(t, err)
	assert.Equal(t, []Command{
		{
			Name: "docker",
			Arguments: []string{
				"pull",
				"us-central1-docker.pkg.dev/gcp-project/repo-name/client-app:latest-green",
			},
		},
		{
			Name: "docker",
			Arguments: []string{
				"tag",
				"us-central1-docker.pkg.dev/gcp-project/repo-name/client-app:latest-green",
				"us-central1-docker.pkg.dev/gcp-project/repo-name/client-app:currentSha",
			},
		},
		{
			Name: "docker",
			Arguments: []string{
				"push",
				"us-central1-docker.pkg.dev/gcp-project/repo-name/client-app:currentSha",
			},
		},
	}, sideEffects.Commands)
}
