package build

func (p Pipeline) ToGitHubWorkflow() GitHubActionsWorkflow {
	workflow := GitHubActionsWorkflow{
		Name: p.Name,
		On: map[string]GitHubActionsTriggerEvent{
			"push": {
				Branches: []string{"trunk"},
			},
		},
		Jobs: p.ArtifactsAndAppsToGitHubActionsJobs(),
	}

	// TODO

	return workflow

}

func (p Pipeline) ArtifactsAndAppsToGitHubActionsJobs() map[string]GitHubActionsJob {
	jobs := map[string]GitHubActionsJob{}

	for id, artifact := range p.artifactsMap {
		jobs["build-"+id] = artifact.ToGitHubActionsJob(p.Cmd, p.ConfigPath)
	}
	for id, app := range p.applicationsMap {
		jobs["deploy-"+id] = app.ToGitHubActionsJob()
	}

	return jobs
}
