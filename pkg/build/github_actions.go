package build

import (
	"os"

	"gopkg.in/yaml.v3"
)

type GitHubActionsWorkflow struct {
	Name string
	On   map[string]GitHubActionsTriggerEvent // v cool https://stackoverflow.com/questions/70849190/golang-how-to-avoid-double-quoted-on-key-on-when-marshaling-struct-to-yaml
	Jobs map[string]GitHubActionsJob
}

func (g GitHubActionsWorkflow) WriteYaml(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(g)
}

type GitHubActionsJob struct {
	Name        string
	RunsOn      string `yaml:"runs-on"`
	Permissions map[string]string
	Needs       []string `yaml:"needs,omitempty"`
	Steps       []GitHubActionsStep
}

type GitHubActionsStep struct {
	Id   string                 `yaml:"id,omitempty"`
	Name string                 `yaml:"name,omitempty"`
	Uses string                 `yaml:"uses,omitempty"`
	With map[string]interface{} `yaml:"with,omitempty"`
	Env  map[string]string      `yaml:"env,omitempty"`
	Run  string                 `yaml:"run,omitempty"`
}

type GitHubActionsTriggerEvent struct {
	Branches []string
}
