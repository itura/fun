package build

import (
	"github.com/itura/fun/pkg/fun"
)

type Job interface {
	JobId() string
}

type PipelineConfig struct {
	Artifacts    fun.Config[Artifact]
	Applications fun.Config[Application]
	Dependencies Dependencies
	BuildName    string
	Error        error
}

func NewParsedConfig() PipelineConfig {
	return PipelineConfig{
		Artifacts:    fun.NewConfig[Artifact](),
		Applications: fun.NewConfig[Application](),
	}
}

func SuccessfulParse(name string, artifacts fun.Config[Artifact], applications fun.Config[Application], dependencies Dependencies) PipelineConfig {
	return NewParsedConfig().
		SetArtifacts(artifacts).
		SetApplications(applications).
		SetBuildName(name).
		SetDependencies(dependencies)
}

func FailedParse(name string, err error) PipelineConfig {
	return NewParsedConfig().
		SetBuildName(name).
		SetError(err)
}

func (c PipelineConfig) SetArtifacts(artifacts fun.Config[Artifact]) PipelineConfig {
	c.Artifacts = artifacts
	return c
}

func (c PipelineConfig) SetApplications(applications fun.Config[Application]) PipelineConfig {
	c.Applications = applications
	return c
}

func (c PipelineConfig) SetDependencies(deps Dependencies) PipelineConfig {
	c.Dependencies = deps
	return c
}

func (c PipelineConfig) SetBuildName(name string) PipelineConfig {
	c.BuildName = name
	return c
}

func (c PipelineConfig) SetError(err error) PipelineConfig {
	c.Error = err
	return c
}
