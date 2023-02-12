package build

import (
	"fmt"
)

type Build interface {
	Build() (SideEffects, error)
}
type NullBuild struct{}

func (b NullBuild) Build() (SideEffects, error) {
	return SideEffects{}, fmt.Errorf("invalid build")
}

type DockerImage struct {
	Artifact
	dockerfile string
	workdir    string
}

func NewDockerImage(a Artifact) DockerImage {
	return DockerImage{
		Artifact:   a,
		dockerfile: fmt.Sprintf("%s/Dockerfile", a.Path),
		workdir:    a.Path,
	}
}

func (b DockerImage) Build() (SideEffects, error) {
	commitTag := b.AppImageName(b.CurrentSha)
	greenTag := b.AppImageName(b.GreenTag())

	if b.hasChanged {
		return NewSideEffects(
			// tests
			NewCommand("docker", "build",
				"-f", b.dockerfile,
				"-t", b.VerifyImageName(),
				"--target", b.VerifyTarget(),
				b.workdir,
			),
			NewCommand("docker", "run", "--rm", b.VerifyImageName()),
			//app
			NewCommand("docker", "build",
				"-f", b.dockerfile,
				"-t", b.AppImageName(b.CurrentSha),
				"--target", b.AppTarget(),
				b.workdir,
			),
			NewCommand("docker", "tag", commitTag, greenTag),
			NewCommand("docker", "push",
				"--all-tags",
				b.AppImageBase(),
			),
		), nil
	} else {
		return NewSideEffects(
			NewCommand("docker", "pull", greenTag),
			NewCommand("docker", "tag", greenTag, commitTag),
			NewCommand("docker", "push", commitTag),
		), nil
	}
}

type HelmDeployment struct {
	Application
}

func NewHelm(a Application) HelmDeployment {
	return HelmDeployment{
		Application: a,
	}
}

func (b HelmDeployment) Build() (SideEffects, error) {
	deploy := NewCommand("helm", "upgrade", b.Id, b.Path,
		"--install",
		"--atomic",
		"--namespace", b.Namespace,
		"--set", fmt.Sprintf("repo=%s", b.Repository),
		"--set", fmt.Sprintf("tag=%s", b.CurrentSha),
	)

	for _, arg := range b.RuntimeArgs {
		deploy = deploy.Add("--set", fmt.Sprintf("%s=$%s", arg.Key, arg.EnvKey()))
	}

	return NewSideEffects(
		NewCommand("helm", "dep", "update"),
		deploy,
	), nil
}

type TfConfig struct {
	Application
}

func NewTerraform(a Application) TfConfig {
	return TfConfig{
		Application: a,
	}
}

func (b TfConfig) Build() (SideEffects, error) {
	chdir := fmt.Sprintf("-chdir=%s", b.Path)
	return NewSideEffects(
		NewCommand("terraform", chdir, "init"),
		NewCommand("terraform", chdir, "plan", "-out=plan.out"),
		NewCommand("terraform", chdir, "apply", "plan.out"),
	), nil
}
