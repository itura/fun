package build

import (
	"github.com/itura/fun/pkg/fun"
	"os"
	"os/exec"
	"strings"
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

type Command interface {
	Run(name string, args ...string) error
	RunSilent(name string, args ...string) error
}

type ShellCommand struct{}

func (c ShellCommand) Run(name string, args ...string) error {
	return command(name, args...)
}

func (c ShellCommand) RunSilent(name string, args ...string) error {
	return commandSilent(name, args...)
}

type MockCommand struct {
}

func (c MockCommand) Run(name string, args ...string) error {
	return command(name, args...)
}

func (c MockCommand) RunSilent(name string, args ...string) error {
	return commandSilent(name, args...)
}

type Job interface {
	JobId() string
}

type ParsedConfig struct {
	Artifacts    fun.Config[Artifact]
	Applications fun.Config[Application]
	BuildName    string
	Error        error
}

func NewParsedConfig() ParsedConfig {
	return ParsedConfig{
		Artifacts:    fun.NewConfig[Artifact](),
		Applications: fun.NewConfig[Application](),
	}
}

func SuccessfulParse(name string, artifacts fun.Config[Artifact], applications fun.Config[Application]) ParsedConfig {
	return NewParsedConfig().
		SetArtifacts(artifacts).
		SetApplications(applications).
		SetBuildName(name)
}

func FailedParse(name string, err error) ParsedConfig {
	return NewParsedConfig().
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
