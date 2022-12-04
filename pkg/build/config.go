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

type ApplicationType string

var (
	typeHelm      ApplicationType = "helm"
	typeTerraform ApplicationType = "terraform"
)

type PipelineConfig struct {
	Name      string
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
		Type         ApplicationType
	}
}

func parseConfig(configPath string, projectId string, currentSha string, previousSha string, force bool) (map[string]Artifact, map[string]Application, string, error) {
	dat, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, "", err
	}

	var config PipelineConfig
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return nil, nil, "", err
	}

	artifacts := make(map[string]Artifact)
	for _, spec := range config.Artifacts {
		var cd ChangeDetection
		if force {
			cd = NewAlwaysChanged()
		} else {
			_cd := NewGitChangeDetection(previousSha).
				AddPaths(spec.Path)

			// todo make agnostic to ordering
			for _, id := range spec.Dependencies {
				_cd = _cd.AddPaths(artifacts[id].Path)
			}
			cd = _cd
		}

		artifacts[spec.Id] = Artifact{
			Type:            spec.Type,
			Id:              spec.Id,
			Path:            spec.Path,
			Project:         projectId,
			CurrentSha:      currentSha,
			hasDependencies: len(spec.Dependencies) > 0,
			hasChanged:      cd.HasChanged(),
		}
	}

	applications := make(map[string]Application)
	for _, spec := range config.Applications {
		var upstreams []Job
		var cd ChangeDetection
		if force {
			cd = NewAlwaysChanged()
		} else {
			_cd := NewGitChangeDetection(previousSha).
				AddPaths(spec.Path)

			// todo make agnostic to ordering
			for _, id := range spec.Artifacts {
				_cd = _cd.AddPaths(artifacts[id].Path)
				upstreams = append(upstreams, artifacts[id])
			}
			for _, id := range spec.Dependencies {
				_cd = _cd.AddPaths(applications[id].Path)
				upstreams = append(upstreams, applications[id])
			}
			cd = _cd
		}

		hasDependencies := len(spec.Dependencies) > 0 || len(spec.Artifacts) > 0
		applications[spec.Id] = Application{
			Type:            spec.Type,
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

	return artifacts, applications, config.Name, nil
}
