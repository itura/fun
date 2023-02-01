package build

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowGeneration(t *testing.T) {
	builder := NewTestBuilder("currentSha")
	expectedYamlBytes, _ := os.ReadFile("test_fixtures/valid_workflow.yaml")
	expectedWorkflow := GitHubActionsWorkflow{}
	err := yaml.Unmarshal(expectedYamlBytes, &expectedWorkflow)
	assert.Nil(t, err)

	apiArtifact := builder.Artifact("api", "packages/api")
	clientArtifact := builder.Artifact("client", "packages/client", apiArtifact)

	dbApp := PostgresHelmChart(builder)
	dbApp.Upstreams = []Job{clientArtifact}
	parsedConfig := SuccessfulParse(
		"My Build",
		map[string]Artifact{
			"api":    apiArtifact,
			"client": clientArtifact,
		},
		map[string]Application{
			"db": dbApp,
		},
	)
	pipeline := NewPipeline(parsedConfig, "test_fixtures/pipeline_config_pass.yaml", "github.com/itura/fun/cmd/build@v0.1.19")

	workflow := pipeline.ToGitHubWorkflow()

	// workflowBytes, _ := yaml.Marshal(workflow)

	// os.WriteFile("test_fixtures/test.yaml", workflowBytes, 0644)

	assert.Equal(t, expectedWorkflow, workflow)
}

func TestDeployTerraformApplication(t *testing.T) {
	builder := NewTestBuilder("currentSha")

	terraformApp := builder.Application("infra", "terraform/main", typeTerraform)
	parsedConfig := SuccessfulParse(
		"My Build",
		map[string]Artifact{},
		map[string]Application{
			"infra": terraformApp,
		},
	)
	pipeline := NewPipeline(parsedConfig, "test_fixtures/pipeline_config_pass.yaml", "github.com/itura/fun/cmd/build@v0.1.19")

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
	t.Skip("MAKE THIS NOT SUCK")
	builder := NewTestBuilder("currentSha")

	dbApp := PostgresHelmChart(builder)
	parsedConfig := SuccessfulParse(
		"My Build",
		map[string]Artifact{},
		map[string]Application{
			"db": dbApp,
		},
	)
	pipeline := NewPipeline(parsedConfig, "test_fixtures/pipeline_config_pass.yaml", "github.com/itura/fun/cmd/build@v0.1.19")

	sideEffects, err := pipeline.DeployApplication("db")

	assert.Nil(t, err)
	assert.Equal(t, sideEffects.Commands, []Command{
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
				"app-namespace",
				"--set",
				fmt.Sprintf("repo=%s", builder.repository()),
				"--set",
				"tag=currentSha",
				"--set",
				"postgresql.dbName=$postgresql_dbName",
				"--set",
				"postgresql.auth.password=$postgresql_auth_password",
				"--set",
				"postgresql.auth.username=$postgresql_auth_password",
			}},
	})

}

// keep the example up to date
func TestPipelineGeneration(t *testing.T) {
	pipeline, err := ParsePipeline(TestArgs("example/pipeline.yaml"), "")
	assert.Nil(t, err)

	err = pipeline.ToGitHubWorkflow().WriteYaml("example/workflow.yaml")
	assert.Nil(t, err)
}
