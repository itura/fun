package build

type SideEffects struct {
	Commands []Command
	Runner   CommandRunner
}

type Command struct {
	Name      string
	Arguments []string
}

func (s SideEffects) Apply() error {
	for _, command := range s.Commands {
		err := s.Runner.Run(command.Name, command.Arguments...)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SideEffects) AddCommand(command Command) {
	s.Commands = append(s.Commands, command)
}

type CommandRunner interface {
	Run(name string, args ...string) error
	RunSilent(name string, args ...string) error
}

type ShellCommandRunner struct{}

func (c ShellCommandRunner) Run(name string, args ...string) error {
	return command(name, args...)
}

func (c ShellCommandRunner) RunSilent(name string, args ...string) error {
	return commandSilent(name, args...)
}
