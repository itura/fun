package build

type SideEffects struct {
	Commands []Command
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
}

type ShellCommandRunner struct{}

func (c ShellCommandRunner) Run(name string, args ...string) error {
	return command(name, args...)
}

func (c ShellCommandRunner) RunSilent(name string, args ...string) error {
	return commandSilent(name, args...)
}
