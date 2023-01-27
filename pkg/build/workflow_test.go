package build

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type TestJob struct {
	Id string
}

func (j TestJob) JobId() string {
	return j.Id
}

func TestWorkflowGeneration(t *testing.T) {
	builder := NewTestBuilder("projectId", "currentSha")
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
