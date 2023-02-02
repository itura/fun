package build

import (
	"fmt"
	"strings"
)

type Build interface {
	Build() (SideEffects, error)
}
type NullBuild struct{}

func (b NullBuild) Build() (SideEffects, error) {
	return SideEffects{}, fmt.Errorf("invalid build")
}

type PackageVerifier struct {
	Artifact
	dockerfile string
	workdir    string
	buildArgs  []string
}

func NewPackageVerifier(a Artifact) PackageVerifier {
	return PackageVerifier{
		Artifact:   a,
		dockerfile: fmt.Sprintf("%s/Dockerfile", a.Path),
		workdir:    a.Path,
	}
}

func (b PackageVerifier) Build() (SideEffects, error) {
	if b.hasChanged {
		return SideEffects{Commands: []Command{
			{
				Name: "docker",
				Arguments: []string{
					"build",
					"-f", b.dockerfile,
					"-t", b.VerifyImageName(),
					"--target", b.VerifyTarget(),
				},
			},
			{
				Name:      "docker",
				Arguments: []string{"run", "--rm", b.VerifyImageName()},
			},
		}}, nil

	}

	return SideEffects{}, nil
}

type DockerImage struct {
	PackageVerifier
}

func NewDockerImage(a Artifact) DockerImage {
	return DockerImage{
		PackageVerifier: NewPackageVerifier(a),
	}
}

func (b DockerImage) Build() (SideEffects, error) {
	sideEffects, err := b.PackageVerifier.Build()
	if err != nil {
		return SideEffects{}, err
	}

	commitTag := b.AppImageName(b.CurrentSha)
	greenTag := b.AppImageName(b.GreenTag())

	if b.hasChanged {
		sideEffects = sideEffects.Add(
			NewCommand("docker", "build",
				"-f", b.dockerfile,
				"-t", b.AppImageName(b.CurrentSha),
				"--target", b.AppTarget(),
			),
			NewCommand("docker", "tag", commitTag, greenTag),
			NewCommand("docker", "push",
				"--all-tags",
				b.AppImageBase(),
			),
		)
	} else {
		sideEffects = sideEffects.Add(
			NewCommand("docker", "pull", greenTag),
			NewCommand("docker", "tag", greenTag, commitTag),
			NewCommand("docker", "push", commitTag),
		)
	}
	return sideEffects, nil
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
	args := []string{
		"upgrade", b.Id, b.Path,
		"--install",
		"--atomic",
		"--namespace", b.Namespace,
		"--set", fmt.Sprintf("repo=%s", b.Repository),
		"--set", fmt.Sprintf("tag=%s", b.CurrentSha),
	}

	for _, value := range b.Values {
		envValue := strings.ReplaceAll(value.Key, ".", "_")
		args = append(args, "--set", fmt.Sprintf("%s=$%s", value.Key, envValue))
	}
	for _, providerSecret := range b.Secrets {
		for _, secret := range providerSecret {
			envVarName := strings.ReplaceAll(secret.HelmKey, ".", "_")
			args = append(args, "--set", fmt.Sprintf("%s=$%s", secret.HelmKey, envVarName))
		}
	}

	return SideEffects{
		Commands: []Command{
			{
				Name: "helm",
				Arguments: []string{
					"dep",
					"update",
				},
			},
			{
				Name:      "helm",
				Arguments: args,
			},
		},
	}, nil
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
	return SideEffects{
		Commands: []Command{
			{
				Name: "terraform",
				Arguments: []string{
					"-chdir=terraform/main",
					"init",
				},
			},
			{
				Name: "terraform",
				Arguments: []string{
					"-chdir=terraform/main",
					"plan",
					"-out=plan.out",
				},
			},
			{
				Name: "terraform",
				Arguments: []string{
					"-chdir=terraform/main",
					"apply",
					"plan.out",
				},
			},
		},
	}, nil
}
