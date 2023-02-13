package build

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed VERSION
var versionFile embed.FS

type Pipeline struct {
	config     PipelineConfig
	ConfigPath string
	Name       string
	Cmd        string
}

func NewPipeline(result PipelineConfig, configPath string, _cmd string) Pipeline {
	return Pipeline{
		config:     result,
		ConfigPath: configPath,
		Name:       result.BuildName,
		Cmd:        _cmd,
	}
}

func ParsePipeline(args ActionArgs, cd ChangeDetection) (Pipeline, error) {
	config := parseConfig(args, cd)
	if config.Error != nil {
		return Pipeline{}, config.Error
	}

	_version, err := versionFile.ReadFile("VERSION")
	if err != nil {
		return Pipeline{}, err
	}
	var _cmd string
	if args.Self {
		_cmd = "cmd/build/main.go"
	} else {
		_cmd = fmt.Sprintf("github.com/itura/fun/cmd/build@%s", _version)
	}

	return NewPipeline(config, args.ConfigPath, _cmd), nil
}

func (p Pipeline) BuildArtifact(id string) (SideEffects, error) {
	artifact, present := p.config.Artifacts[id]
	if !present {
		return SideEffects{}, fmt.Errorf("invalid id %s", id)
	}
	build, err := artifact.PrepareBuild()
	if err != nil {
		return SideEffects{}, err
	}
	return build.Build()
}

func (p Pipeline) DeployApplication(id string) (SideEffects, error) {
	application, present := p.config.Applications[id]
	if !present {
		return SideEffects{}, fmt.Errorf("invalid id %s", id)
	}
	build := application.PrepareBuild()

	return build.Build()
}

func (p Pipeline) ToGitHubWorkflow() GitHubActionsWorkflow {
	jobs := map[string]GitHubActionsJob{}

	dependencies := p.config.Dependencies
	for id, artifact := range p.config.Artifacts {
		jobId := dependencies.GetJobId(id)
		jobs[jobId] = artifact.ToGitHubActionsJob(p.Cmd, p.ConfigPath, dependencies)
	}
	for id, app := range p.config.Applications {
		jobId := dependencies.GetJobId(id)
		jobs[jobId] = app.ToGitHubActionsJob(p.Cmd, p.ConfigPath, dependencies)
	}

	workflow := GitHubActionsWorkflow{
		Name: p.Name,
		On: map[string]GitHubActionsTriggerEvent{
			"push": {
				Branches: []string{"trunk"},
			},
		},
		Jobs: jobs,
	}

	return workflow
}

func resolveKey(value RuntimeArg) string {
	return strings.ReplaceAll(value.Key, ".", "_")
}

func formatEnvValue(name string) string {
	return fmt.Sprintf("${{ env.%s }}", name)
}

func formatSecretValue(name string) string {
	return fmt.Sprintf("${{ secrets.%s }}", name)
}
