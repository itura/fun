package build

import (
	"encoding/json"
	"fmt"
	"github.com/alexflint/go-arg"
	"os"
)

type argv struct {
	BuildArtifact     *BuildArtifactCommand     `arg:"subcommand:build-artifact"`
	DeployApplication *DeployApplicationCommand `arg:"subcommand:deploy-application"`
	Generate          *GenerateCommand          `arg:"subcommand:generate"`
}

func (a argv) Version() string {
	version, err := versionFile.ReadFile("VERSION")
	if err != nil {
		return ""
	} else {
		return string(version)
	}
}

func Run() int {
	var args argv
	cli := arg.MustParse(&args)
	command, ok := cli.Subcommand().(PipelineCommand)
	if !ok {
		cli.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	data, err := json.MarshalIndent(&command, "", "  ")
	if err != nil {
		return bail(err)
	}
	fmt.Printf("fun/build %s using %s\n", args.Version(), data)

	err = command.Run()
	if err != nil {
		return bail(err)
	}

	fmt.Printf("😎\n")
	return 0
}

func bail(err error) int {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	fmt.Fprintf(os.Stderr, "😭\n")
	return 1
}
