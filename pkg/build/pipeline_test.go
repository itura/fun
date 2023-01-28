package build

import (
	"gopkg.in/yaml.v3"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowGeneration(t *testing.T) {
	builder := NewTestBuilder("", "currentSha")
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

// keep the example up to date
func TestPipelineGeneration(t *testing.T) {
	pipeline, err := ParsePipeline(TestArgs("example/pipeline.yaml"), "")
	assert.Nil(t, err)

	err = pipeline.ToGitHubWorkflow().WriteYaml("example/workflow.yaml")
	assert.Nil(t, err)
}
