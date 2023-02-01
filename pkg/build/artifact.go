package build

import (
	"fmt"
)

type Artifact struct {
	Id              string
	Path            string
	Repository      string
	Host            string
	CurrentSha      string
	Type            ArtifactType
	Upstreams       []Job
	hasDependencies bool
	hasChanged      bool
	CloudProvider   CloudProviderConfig
}

func CreateArtifacts(args ActionArgs, previousSha string, config PipelineConfigRaw, repository string) map[string]Artifact {
	artifacts := make(map[string]Artifact)
	for _, spec := range config.Artifacts {
		var upstreams []Job
		var cd ChangeDetection
		if args.Force {
			cd = NewAlwaysChanged()
		} else {
			_cd := NewGitChangeDetection(previousSha).
				AddPaths(spec.Path)

			// todo make agnostic to ordering
			for _, id := range spec.Dependencies {
				_cd = _cd.AddPaths(artifacts[id].Path)
				upstreams = append(upstreams, artifacts[id])
			}
			cd = _cd
		}

		artifacts[spec.Id] = Artifact{
			Type:            spec.Type,
			Id:              spec.Id,
			Path:            spec.Path,
			Repository:      repository,
			Host:            config.Resources.ArtifactRepository.Host,
			CurrentSha:      args.CurrentSha,
			hasDependencies: len(spec.Dependencies) > 0,
			Upstreams:       upstreams,
			hasChanged:      cd.HasChanged(),
			CloudProvider:   config.Resources.CloudProvider,
		}
	}
	return artifacts
}

func (a Artifact) PrepareBuild() (Build, error) {
	switch a.Type {
	case typeLib:
		return NewPackageVerifier(a), nil
	case typeApp:
		return NewDockerImage(a), nil
	default:
		return NullBuild{}, fmt.Errorf("invalid artifact type %s", a.Type)
	}
}

func (a Artifact) GreenTag() string {
	return "latest-green"
}

func (a Artifact) JobId() string {
	return fmt.Sprintf("build-%s", a.Id)
}

func (a Artifact) VerifyTarget() string {
	return "test"
}

func (a Artifact) VerifyImageName() string {
	return fmt.Sprintf("%s/%s-%s:%s", a.Repository, a.Id, a.VerifyTarget(), a.CurrentSha)
}

func (a Artifact) AppTarget() string {
	return "app"
}

func (a Artifact) AppImageBase() string {
	return fmt.Sprintf("%s/%s-%s", a.Repository, a.Id, a.AppTarget())
}

func (a Artifact) AppImageName(tag string) string {
	return fmt.Sprintf("%s:%s", a.AppImageBase(), tag)
}

func (a Artifact) HasDependencies() bool {
	return a.hasDependencies
}
