package build

import (
	"os"
	"testing"
)

var steps = []GitHubActionsStep{
	{
		Id:   "",
		Uses: "google-github-actions/get-gke-credentials@v1",
		With: map[string]string{
			"cluster_name": "cluster-name",
			"location":     "uscentral1",
		},
	},
	{
		Id:   "",
		Uses: "azure/setup-helm@v3",
		With: map[string]string{
			"version": "v3.10.2",
		},
	},
	{
		Id:   "secrets",
		Uses: "google-github-actions/get-secretmanager-secrets@v1",
		With: map[string]string{
			"secrets": `
pg-password:princess-pup/pg-password`,
		},
	},
}

func TestApplicationSetup(t *testing.T) {
	app := getHelmApplication()
	expectedStepsString, _ := MarshalAndIndentSteps(steps)

	stepsStr := app.Setup()

	if stepsStr != expectedStepsString {
		t.Fatalf("Expected setup steps to be \n`%s`\nBut got \n`%s`\ninstead", expectedStepsString, stepsStr)
	}
}

func TestMarshalAndIndentSteps(t *testing.T) {
	// load from file due to immense difficulty with multiline strings & preserving indents
	expectedYamlBytes, err := os.ReadFile("test_fixtures/sample_setup_actions.txt")
	if err != nil {
		t.Fatalf("Error loading test fixture: %s", err)
	}

	expectedYamlStr := string(expectedYamlBytes)

	stepsYamlStr, err := MarshalAndIndentSteps(steps)
	if err != nil {
		t.Fatalf("Error while marshaling YAML: %s", err)
	}

	if stepsYamlStr != expectedYamlStr {
		t.Fatalf("Expected marshaled GHA steps to be \n`%s`\nBut got \n`%s`\ninstead", expectedYamlStr, stepsYamlStr)
	}

}
