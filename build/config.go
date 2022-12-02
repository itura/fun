package build

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Preset string

var (
	presetGolang Preset = "golang"
)

type ArtifactType string

var (
	typeLibGo ArtifactType = "lib-go"
	typeAppGo ArtifactType = "app-go"
	typeApp   ArtifactType = "app"
	typeLib   ArtifactType = "lib"
)

type PipelineConfig struct {
	Artifacts []struct {
		Id           string
		Path         string
		Dependencies []string
		Type         ArtifactType
	}
	Applications []struct {
		Id           string
		Path         string
		Namespace    string
		Artifacts    []string
		Values       []HelmValue
		Dependencies []string
	}
}

func parseConfig(configPath string, projectId string, currentSha string, previousSha string) (map[string]Artifact, map[string]Application, error) {
	dat, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, err
	}

	var config PipelineConfig
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return nil, nil, err
	}

	artifacts := make(map[string]Artifact)
	for _, spec := range config.Artifacts {
		cd := NewGitChangeDetection(previousSha).
			AddPaths(spec.Path)

		// todo make agnostic to ordering
		for _, id := range spec.Dependencies {
			cd = cd.AddPaths(artifacts[id].Path)
		}

		artifacts[spec.Id] = Artifact{
			Id:              spec.Id,
			Path:            spec.Path,
			Project:         projectId,
			CurrentSha:      currentSha,
			Type:            spec.Type,
			hasDependencies: len(spec.Dependencies) > 0,
			hasChanged:      cd.HasChanged(),
		}
	}

	applications := make(map[string]Application)
	for _, spec := range config.Applications {
		cd := NewGitChangeDetection(previousSha).
			AddPaths(spec.Path)
		var upstreams []Job

		// todo make agnostic to ordering
		for _, id := range spec.Artifacts {
			cd = cd.AddPaths(artifacts[id].Path)
			upstreams = append(upstreams, artifacts[id])
		}
		for _, id := range spec.Dependencies {
			cd = cd.AddPaths(applications[id].Path)
			upstreams = append(upstreams, applications[id])
		}

		hasDependencies := len(spec.Dependencies) > 0 || len(spec.Artifacts) > 0
		applications[spec.Id] = Application{
			Id:              spec.Id,
			Path:            spec.Path,
			ProjectId:       projectId,
			CurrentSha:      currentSha,
			Namespace:       spec.Namespace,
			Values:          spec.Values,
			Upstreams:       upstreams,
			hasDependencies: hasDependencies,
			hasChanged:      cd.HasChanged(),
		}
	}

	return artifacts, applications, nil
}
