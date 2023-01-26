package build

type GitHubActionsWorkflow struct {
	Name string
	On   map[string]GitHubActionsTriggerEvent
	Jobs map[string]GitHubActionsJob
}

type GitHubActionsJob struct {
	Name        string
	RunsOn      string `yaml:"runs-on"`
	Permissions GitHubActionsJobPermissions
	Needs       []string
	Steps       []GitHubActionsStep
}

type GitHubActionsJobPermissions struct {
	IdToken  string `yaml:"id-token"`
	Contents string
}

type GitHubActionsStep struct {
	Id   string `yaml:"id,omitempty"`
	Uses string
	With map[string]string
}

type GitHubActionsTriggerEvent struct {
	Branches []string
}
