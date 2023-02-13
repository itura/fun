package build

import (
	"os"
	"os/exec"
	"strings"
)

type SideEffects struct {
	Commands []Command
}

func NewSideEffects(commands ...Command) SideEffects {
	return SideEffects{
		Commands: commands,
	}
}

type Command struct {
	Name      string
	Arguments []string
}

func NewCommand(name string, args ...string) Command {
	return Command{
		Name:      name,
		Arguments: args,
	}
}

func (c Command) Add(args ...string) Command {
	c.Arguments = append(c.Arguments, args...)
	return c
}

func (s SideEffects) Apply(r CommandRunner) error {
	for _, command := range s.Commands {
		err := r.Run(command.Name, command.Arguments...)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s SideEffects) Add(commands ...Command) SideEffects {
	s.Commands = append(s.Commands, commands...)
	return s
}

type CommandRunner interface {
	Run(name string, args ...string) error
	RunSilent(name string, args ...string) error
	Output(name string, args ...string) (string, error)
}

type ShellCommandRunner struct{}

func (c ShellCommandRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c ShellCommandRunner) RunSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	_, err := cmd.Output()
	return err
}

func (c ShellCommandRunner) Output(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	raw, err := cmd.Output()
	result := strings.TrimSpace(string(raw))
	return result, err
}
