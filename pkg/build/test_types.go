package build

type InputArguments struct {
	ProjectId   string
	CurrentSha  string
	PreviousSha string
	Force       bool
}

type ParseConfigOutputs struct {
	Artifacts    map[string]Artifact
	Applications map[string]Application
	BuildName    string
	Error        error
}
