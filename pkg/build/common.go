package build

import (
	"os"
	"os/exec"
	"strings"

	"github.com/itura/fun/pkg/fun"
)

func command(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func commandSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	_, err := cmd.Output()
	return err
}

func previousCommit() string {
	revList := exec.Command("git", "rev-list", "-n", "1", "HEAD~1")
	output, _ := revList.Output()
	ref := strings.TrimSpace(string(output))
	return ref
}

type Job interface {
	JobId() string
}

type PipelineConfig struct {
	Artifacts    fun.Config[Artifact]
	Applications fun.Config[Application]
	BuildName    string
	Error        error
}

func NewParsedConfig() PipelineConfig {
	return PipelineConfig{
		Artifacts:    fun.NewConfig[Artifact](),
		Applications: fun.NewConfig[Application](),
	}
}

func SuccessfulParse(name string, artifacts fun.Config[Artifact], applications fun.Config[Application]) PipelineConfig {
	return NewParsedConfig().
		SetArtifacts(artifacts).
		SetApplications(applications).
		SetBuildName(name)
}

func FailedParse(name string, err error) PipelineConfig {
	return NewParsedConfig().
		SetBuildName(name).
		SetError(err)
}

func (o PipelineConfig) ListArtifacts() []Artifact {
	var artifacts []Artifact
	for _, v := range o.Artifacts {
		artifacts = append(artifacts, v)
	}
	return artifacts
}

func (o PipelineConfig) ListApplications() []Application {
	var applications []Application
	for _, v := range o.Applications {
		applications = append(applications, v)
	}
	return applications
}

func (o PipelineConfig) SetArtifacts(artifacts fun.Config[Artifact]) PipelineConfig {
	o.Artifacts = artifacts
	return o
}

func (o PipelineConfig) SetApplications(applications fun.Config[Application]) PipelineConfig {
	o.Applications = applications
	return o
}

func (o PipelineConfig) SetBuildName(name string) PipelineConfig {
	o.BuildName = name
	return o
}

func (o PipelineConfig) SetError(err error) PipelineConfig {
	o.Error = err
	return o
}
