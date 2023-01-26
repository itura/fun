package build

type GitHubActionsWorkflow struct {
	Name string
	On   map[string]GitHubActionsTriggerEvent
	Jobs map[string]GitHubActionsJob
}

type GitHubActionsJob struct {
	Name        string
	RunsOn      string `yaml:"runs-on"`
	Permissions map[string]string
	Needs       []string `yaml:"needs,omitempty"`
	Steps       []GitHubActionsStep
}

type GitHubActionsStep struct {
	Id   string            `yaml:"id,omitempty"`
	Name string            `yaml:"name,omitempty"`
	Uses string            `yaml:"uses,omitempty"`
	With map[string]string `yaml:"with,omitempty"`
	Env  map[string]string `yaml:"env,omitempty"`
	Run  string            `yaml:"run,omitempty"`
}

type GitHubActionsTriggerEvent struct {
	Branches []string
}
