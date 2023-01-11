package build

import (
	"fmt"
	"os"
)

type Build interface {
	Build() error
}

type NullBuild struct{}

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

func NewGoPackageVerifier(a Artifact) PackageVerifier {
	return PackageVerifier{
		Artifact:   a,
		dockerfile: "Dockerfile-go",
		workdir:    ".",
		buildArgs: []string{
			fmt.Sprintf("APP_DIR=%s", a.Path),
		},
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

func NewGoDockerImage(a Artifact) DockerImage {
	return DockerImage{
		PackageVerifier: NewGoPackageVerifier(a),
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
		"--timeout", "1m",
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
