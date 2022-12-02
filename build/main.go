package build

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"os"
)

var argv struct {
	BuildArtifact     *BuildArtifactCommand     `arg:"subcommand:build-artifact"`
	DeployApplication *DeployApplicationCommand `arg:"subcommand:deploy-application"`
	Generate          *GenerateCommand          `arg:"subcommand:generate"`
}

func Run() int {
	cli := arg.MustParse(&argv)
	command, ok := cli.Subcommand().(PipelineCommand)
	if !ok {
		cli.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	err := command.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		fmt.Fprintf(os.Stderr, "😭\n")
		return 1
	}

	fmt.Printf("😎\n")
	return 0
}
