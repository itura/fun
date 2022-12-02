package build

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"strings"
	"text/template"
)

//go:embed ci-cd.yaml.tmpl
var ghActionsTemplate embed.FS

//go:embed VERSION
var versionFile embed.FS

type Pipeline struct {
	artifactsMap    map[string]Artifact
	applicationsMap map[string]Application
	ConfigPath      string
	Artifacts       []Artifact
	Applications    []Application
	Name            string
	Cmd             string
}

func NewPipeline(configPath string, projectId string, currentSha string, previousSha string, force bool, self bool) (Pipeline, error) {
	artifactsMap, applicationsMap, name, err := parseConfig(configPath, projectId, currentSha, previousSha, force)
	if err != nil {
		return Pipeline{}, err
	}

	var artifacts []Artifact
	for _, v := range artifactsMap {
		artifacts = append(artifacts, v)
	}
	var applications []Application
	for _, v := range applicationsMap {
		applications = append(applications, v)
	}

	_version, err := versionFile.ReadFile("VERSION")
	if err != nil {
		return Pipeline{}, err
	}
	var _cmd string
	if self {
		_cmd = "cmd/build/main.go"
	} else {
		_cmd = fmt.Sprintf("github.com/itura/fun/cmd/build@%s", _version)
	}

	return Pipeline{
		ConfigPath:      configPath,
		artifactsMap:    artifactsMap,
		Artifacts:       artifacts,
		applicationsMap: applicationsMap,
		Applications:    applications,
		Name:            name,
		Cmd:             _cmd,
	}, nil
}

func (p Pipeline) BuildArtifact(id string) error {
	artifact, present := p.artifactsMap[id]
	if !present {
		return fmt.Errorf("invalid id %s", id)
	}
	build, err := artifact.PrepareBuild()
	if err != nil {
		return err
	}
	return build.Build()
}

func (p Pipeline) DeployApplication(id string) error {
	application, present := p.applicationsMap[id]
	if !present {
		return fmt.Errorf("invalid id %s", id)
	}
	build, err := application.PrepareBuild()
	if err != nil {
		return nil
	}
	return build.Build()
}

func (p Pipeline) ToGithubActions(outputPath string) error {
	tplData, err := ghActionsTemplate.ReadFile("ci-cd.yaml.tmpl")
	if err != nil {
		return err
	}

	tpl, err := template.New("workflow").
		Funcs(template.FuncMap{
			"secret":       formatSecretValue,
			"env":          formatEnvValue,
			"resolveValue": resolveValue,
			"resolveKey":   resolveKey,
		}).
		Parse(string(tplData))
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	err = tpl.Execute(&buffer, p)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, buffer.Bytes(), 0644)
}

func resolveValue(value HelmValue) string {
	if value.Value != "" {
		return value.Value
	} else if value.SecretValue != "" {
		return formatSecretValue(value.SecretValue)
	} else if value.EnvValue != "" {
		return formatEnvValue(value.SecretValue)
	} else {
		return ""
	}
}

func resolveKey(value HelmValue) string {
	return strings.ReplaceAll(value.Key, ".", "_")
}

func formatEnvValue(name string) string {
	return fmt.Sprintf("${{ env.%s }}", name)
}

func formatSecretValue(name string) string {
	return fmt.Sprintf("${{ secrets.%s }}", name)
}
