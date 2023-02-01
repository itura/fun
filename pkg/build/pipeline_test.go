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

	piplineConfig := ValidPipelineConfig(builder)
	pipeline := NewPipeline(piplineConfig, "test_fixtures/valid_pipeline_config.yaml", "github.com/itura/fun/cmd/build@v0.1.22")

	workflow := pipeline.ToGitHubWorkflow()

	assert.Equal(t, expectedWorkflow, workflow)
}

func TestWorkflowGenerationE2e(t *testing.T) {
	expectedYamlBytes, _ := os.ReadFile("test_fixtures/valid_workflow.yaml")
	expectedWorkflow := GitHubActionsWorkflow{}
	err := yaml.Unmarshal(expectedYamlBytes, &expectedWorkflow)
	assert.Nil(t, err)

	pipeline, err := ParsePipeline(TestArgs("test_fixtures/valid_pipeline_config.yaml"), "prevSha")
	assert.Nil(t, err)
	assert.Equal(t, expectedWorkflow, pipeline.ToGitHubWorkflow())

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
	pipeline := NewPipeline(parsedConfig, "test_fixtures/valid_pipeline_config.yaml", "github.com/itura/fun/cmd/build@v0.1.19")

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
