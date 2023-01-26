package build

func (p Pipeline) ToGitHubWorkflow() GitHubActionsWorkflow {
	workflow := GitHubActionsWorkflow{
		Name: p.Name,
		On: map[string]GitHubActionsTriggerEvent{
			"push": {
				Branches: []string{"trunk"},
			},
		},
	}

	// TODO

	return workflow

}
