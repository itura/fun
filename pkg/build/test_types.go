package build

import "github.com/itura/fun/pkg/fun"

type InputArguments struct {
	ProjectId   string
	CurrentSha  string
	PreviousSha string
	Force       bool
}

type ParsedConfig struct {
	Artifacts    fun.Config[Artifact]
	Applications fun.Config[Application]
	BuildName    string
	Error        error
}

func NewParseConfigOutputs() ParsedConfig {
	return ParsedConfig{
		Artifacts:    fun.NewConfig[Artifact](),
		Applications: fun.NewConfig[Application](),
	}
}

func SuccessfulParse(name string, artifacts fun.Config[Artifact], applications fun.Config[Application]) ParsedConfig {
	return NewParseConfigOutputs().
		SetArtifacts(artifacts).
		SetApplications(applications).
		SetBuildName(name)
}

func FailedParse(name string, err error) ParsedConfig {
	return NewParseConfigOutputs().
		SetBuildName(name).
		SetError(err)
}

func (o ParsedConfig) ListArtifacts() []Artifact {
	var artifacts []Artifact
	for _, v := range o.Artifacts {
		artifacts = append(artifacts, v)
	}
	return artifacts
}

func (o ParsedConfig) ListApplications() []Application {
	var applications []Application
	for _, v := range o.Applications {
		applications = append(applications, v)
	}
	return applications
}

func (o ParsedConfig) SetArtifacts(artifacts fun.Config[Artifact]) ParsedConfig {
	o.Artifacts = artifacts
	return o
}

func (o ParsedConfig) SetApplications(applications fun.Config[Application]) ParsedConfig {
	o.Applications = applications
	return o
}

func (o ParsedConfig) SetBuildName(name string) ParsedConfig {
	o.BuildName = name
	return o
}

func (o ParsedConfig) SetError(err error) ParsedConfig {
	o.Error = err
	return o
}
