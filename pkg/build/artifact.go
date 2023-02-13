package build

import (
	"fmt"
)

type Artifact struct {
	Id            string
	Path          string
	Repository    string
	Host          string
	CurrentSha    string
	hasChanged    bool
	CloudProvider CloudProviderConfig
}

func CreateArtifacts(args ActionArgs, cd ChangeDetection, config PipelineConfigRaw, artifactRepository string) map[string]Artifact {
	artifacts := make(map[string]Artifact)
	for _, spec := range config.Artifacts {
		artifacts[spec.Id] = Artifact{
			Id:            spec.Id,
			Path:          spec.Path,
			Repository:    artifactRepository,
			Host:          config.Resources.ArtifactRepository.Host,
			CurrentSha:    args.CurrentSha,
			hasChanged:    cd.HasChanged(spec.Path),
			CloudProvider: config.Resources.CloudProvider,
		}
	}
	return artifacts
}

func (a Artifact) PrepareBuild() (Build, error) {
	return NewDockerImage(a), nil
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
