package build

import (
	"fmt"
)

type Artifact struct {
	Id         string
	Path       string
	Project    string
	CurrentSha string
	Type       ArtifactType

	hasDependencies bool
	hasChanged      bool
}

func (a Artifact) PrepareBuild() (Build, error) {
	switch a.Type {
	case typeLib:
		return NewPackageVerifier(a), nil
	case typeApp:
		return NewDockerImage(a), nil
	case typeLibGo:
		return NewGoPackageVerifier(a), nil
	case typeAppGo:
		return NewGoDockerImage(a), nil
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
	return fmt.Sprintf("gcr.io/%s/%s-%s:%s", a.Project, a.Id, a.VerifyTarget(), a.CurrentSha)
}

func (a Artifact) AppTarget() string {
	return "app"
}

func (a Artifact) AppImageBase() string {
	return fmt.Sprintf("gcr.io/%s/%s-%s", a.Project, a.Id, a.AppTarget())
}

func (a Artifact) AppImageName(tag string) string {
	return fmt.Sprintf("%s:%s", a.AppImageBase(), tag)
}

func (a Artifact) HasDependencies() bool {
	return a.hasDependencies
}
