package build

import (
	"os"
	"os/exec"
	"strings"
)

func command(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func commandSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	_, err := cmd.Output()
	return err
}

func previousCommit() string {
	revList := exec.Command("git", "rev-list", "-n", "1", "HEAD~1")
	output, _ := revList.Output()
	ref := strings.TrimSpace(string(output))
	return ref
}

type Command interface {
	Run(name string, args ...string) error
	RunSilent(name string, args ...string) error
}

type ShellCommand struct{}

func (c ShellCommand) Run(name string, args ...string) error {
	return command(name, args...)
}

func (c ShellCommand) RunSilent(name string, args ...string) error {
	return commandSilent(name, args...)
}

type MockCommand struct {
}

func (c MockCommand) Run(name string, args ...string) error {
	return command(name, args...)
}

func (c MockCommand) RunSilent(name string, args ...string) error {
	return commandSilent(name, args...)
}

type Job interface {
	JobId() string
}
