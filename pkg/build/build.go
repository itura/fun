package build

import (
	"fmt"
	"os"
	"strings"
)

type Build interface {
	Build() error
}

type Build1 interface {
	Build1() (SideEffects, error)
}
type NullBuild struct{}

func (b NullBuild) Build1() (SideEffects, error) {
	return SideEffects{}, fmt.Errorf("invalid build")
}

func (b NullBuild) Build() error {
	return fmt.Errorf("invalid build")
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

func (b PackageVerifier) Build() error {
	if b.hasChanged {
		if err := b.buildImage(b.VerifyImageName(), b.VerifyTarget()); err != nil {
			return err
		}
		if err := b.verify(); err != nil {
			return err
		}
	}
	return nil
}

func (b PackageVerifier) Build1() (SideEffects, error) {
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

func (b PackageVerifier) buildImage(imageTag string, target string) error {
	fmt.Printf("Building %s image %s\n", target, imageTag)
	args := []string{"build",
		"-f", b.dockerfile,
		"-t", imageTag,
		"--target", target,
	}
	for _, arg := range b.buildArgs {
		args = append(args, "--build-arg", arg)
	}
	args = append(args, b.workdir)
	if err := command("docker", args...); err != nil {
		return fmt.Errorf("Failed to build %s image %s\n", target, imageTag)
	}
	return nil
}

func (b PackageVerifier) verify() error {
	fmt.Printf("Verifying %s\n", b.Id)
	err := command("docker", "run", "--rm", b.VerifyImageName())
	if err != nil {
		return fmt.Errorf("Failed to verify build for %s\n", b.Id)
	}
	return nil
}

type DockerImage struct {
	PackageVerifier
}

func NewDockerImage(a Artifact) DockerImage {
	return DockerImage{
		PackageVerifier: NewPackageVerifier(a),
	}
}

func (b DockerImage) Build() error {
	if err := b.PackageVerifier.Build(); err != nil {
		return err
	}
	if b.hasChanged {
		if err := b.buildImage(b.AppImageName(b.CurrentSha), b.AppTarget()); err != nil {
			return err
		}
		if err := b.pushAppImage(); err != nil {
			return err
		}
	} else {
		if err := b.pushAppImageAlias(); err != nil {
			return err
		}
	}
	return nil
}

func (b DockerImage) Build1() (SideEffects, error) {
	sideEffects, err := b.PackageVerifier.Build1()
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

func (b DockerImage) pushAppImage() error {
	commitTag := b.AppImageName(b.CurrentSha)
	greenTag := b.AppImageName(b.GreenTag())

	fmt.Printf("Pushing image %s\n", commitTag)
	err := command("docker", "tag", commitTag, greenTag)
	if err == nil {
		err = command("docker", "push", "--all-tags", b.AppImageBase())
	}

	if err != nil {
		return fmt.Errorf("Failed to push image %s\n", commitTag)
	}
	return nil
}

func (b DockerImage) pushAppImageAlias() error {
	current := b.AppImageName(b.CurrentSha)
	previous := b.AppImageName(b.GreenTag())

	fmt.Printf("Moving %s tag to %s\n", b.GreenTag(), current)
	err := command("docker", "pull", previous)
	if err == nil {
		err = command("docker", "tag", previous, current)
		if err == nil {
			err = command("docker", "push", current)
		}
	}
	if err != nil {
		return fmt.Errorf("Failed to move %s tag to %s\n", b.GreenTag(), current)
	}
	return nil
}

type HelmDeployment struct {
	Application
}

func NewHelm(a Application) HelmDeployment {
	return HelmDeployment{
		Application: a,
	}
}

func (b HelmDeployment) Build1() (SideEffects, error) {
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

func (b HelmDeployment) Build() error {
	if b.hasChanged {
		return b.deployApplication()
	} else {
		fmt.Println("Nothing to deploy")
		return nil
	}
}

func (b HelmDeployment) deployApplication() error {
	err := command("helm", "dep", "update", b.Path)
	if err != nil {
		return err
	}

	args := []string{
		"upgrade", b.Id, b.Path,
		"--install",
		"--atomic",
		"--namespace", b.Namespace,
		"--set", fmt.Sprintf("repo=%s", b.Repository),
		"--set", fmt.Sprintf("tag=%s", b.CurrentSha),
	}
	for _, value := range b.Values {
		envValue := os.Getenv(resolveKey(value))
		args = append(args, "--set", fmt.Sprintf("%s=%s", value.Key, envValue))
	}
	return command("helm", args...)
}

type TfConfig struct {
	Application
}

func NewTerraform(a Application) TfConfig {
	return TfConfig{
		Application: a,
	}
}

func (b TfConfig) Build() error {
	err := command("terraform", fmt.Sprintf("-chdir=%s", b.Path), "init")
	if err != nil {
		return err
	}

	err = command("terraform", fmt.Sprintf("-chdir=%s", b.Path), "plan", "-out=plan.out")
	if err != nil {
		return err
	}

	err = command("terraform", fmt.Sprintf("-chdir=%s", b.Path), "apply", "plan.out")
	if err != nil {
		return err
	}

	return nil
}

func (b TfConfig) Build1() (SideEffects, error) {
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
