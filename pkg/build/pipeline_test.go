package build

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// might be hard to get this to pass, but we should be able
// to get it to a point where only stuff like ordering is off,
// then remove the assertion.
// could consider keeping the test to just generate the example whenever we run verify
func TestPipelineGeneration(t *testing.T) {
	t.Skip("delete this line if you like large diffs")

	// init
	pipeline, err := ParsePipeline(TestArgs("example/pipeline.yaml"), "")
	assert.Nil(t, err)

	// write files
	err = pipeline.ToGithubActions("example/workflow.yaml")
	assert.Nil(t, err)

	err = pipeline.ToGitHubWorkflow().WriteYaml("example/workflow1.yaml")
	assert.Nil(t, err)

	// read files
	original, err := os.ReadFile("example/workflow.yaml")
	assert.Nil(t, err)
	updated, err := os.ReadFile("example/workflow1.yaml")
	assert.Nil(t, err)

	// compare
	assert.Equal(t, string(original), string(updated))
}
