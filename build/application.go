package build

import (
	"fmt"
)

type HelmValue struct {
	Key         string
	Value       string
	SecretValue string `yaml:"secretValue"`
	EnvValue    string `yaml:"envValue"`
}

type Application struct {
	Id         string
	Path       string
	ProjectId  string
	CurrentSha string
	Namespace  string
	Values     []HelmValue
	Upstreams  []Job

	hasDependencies bool
	hasChanged      bool
}

func (a Application) PrepareBuild() (Build, error) {
	return NewHelmDeployment(a), nil
}

func (a Application) JobId() string {
	return fmt.Sprintf("deploy-%s", a.Id)
}

func (a Application) HasDependencies() bool {
	return a.hasDependencies
}
