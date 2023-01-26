package build

func (p Pipeline) ToGitHubWorkflow() GitHubActionsWorkflow {
	workflow := GitHubActionsWorkflow{
		Name: p.Name,
		On: map[string]GitHubActionsTriggerEvent{
			"push": {
				Branches: []string{"trunk"},
			},
		},
		Jobs: p.ArtifactsMapToGitHubActionsJobs(),
	}

	// TODO

	return workflow

}

func (p Pipeline) ArtifactsMapToGitHubActionsJobs() map[string]GitHubActionsJob {
	jobs := map[string]GitHubActionsJob{}
	for id, artifact := range p.artifactsMap {
		jobs["build-"+id] = artifact.ToGitHubActionsJob()
	}
	return jobs
}
