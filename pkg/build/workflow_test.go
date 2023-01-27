package build

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestWorkflowGeneration(t *testing.T) {
	builder := NewTestBuilder("projectId", "currentSha")
	expectedYamlBytes, _ := os.ReadFile("test_fixtures/valid_workflow.yaml")
	expectedWorkflow := GitHubActionsWorkflow{}
	err := yaml.Unmarshal(expectedYamlBytes, &expectedWorkflow)
	assert.Nil(t, err)

	parsedConfig := SuccessfulParse(
		"My Build",
		map[string]Artifact{
			"api": builder.Artifact("api", "packages/api"),
		},
		map[string]Application{
			"db": PostgresHelmChart(builder),
		},
	)
	pipeline := NewPipeline(parsedConfig, "test_fixtures/pipeline_config_pass.yaml", "github.com/itura/fun/cmd/build@v0.1.19")

	workflow := pipeline.ToGitHubWorkflow()

	// workflowBytes, _ := yaml.Marshal(workflow)

	// os.WriteFile("test_fixtures/test.yaml", workflowBytes, 0644)

	assert.Equal(t, expectedWorkflow, workflow)
}
