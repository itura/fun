package build

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestWorkflowGeneration(t *testing.T) {
	expectedYamlBytes, _ := os.ReadFile("test_fixtures/valid_workflow.yaml")
	expectedWorkflow := GitHubActionsWorkflow{
		Name: "My Build",
		On: map[string]GitHubActionsTriggerEvent{
			"push": {
				Branches: []string{
					"main",
				},
			},
		},
		Jobs: map[string]GitHubActionsJob{
			"build-api": getAppArtifact().ToGitHubActionsJob(),
		},
	}
	err := yaml.Unmarshal(expectedYamlBytes, &expectedWorkflow)
	assert.Nil(t, err)

	parsedConfig := SuccessfulParse("My Build", getValidArtifacts(), getValidApplications())
	pipeline := NewPipeline(parsedConfig, "test_fixtures/pipeline_config_pass.yaml", "generate")

	workflow := pipeline.ToGitHubWorkflow()

	assert.Equal(t, expectedWorkflow, workflow)
}
