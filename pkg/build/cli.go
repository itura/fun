package build

type PipelineCommand interface {
	Run() error
}

type CommonArgs struct {
	ConfigPath string `arg:"--config" default:"pipeline.yaml" help:"path to pipeline definition yaml"`
	Self       bool   `arg:"--self" help:"Run the tool in its own source repo."`
}

type ActionArgs struct {
	CommonArgs
	Id         string `arg:"positional,required"`
	CurrentSha string `arg:"--current-sha,required" help:"current git sha, used for change detection"`
	Force      bool   `arg:"--force" help:"Ignore change detection"`
}

func (a ActionArgs) CreatePipeline() (Pipeline, error) {
	return ParsePipeline(a, previousCommit())
}

type GenerateArgs struct {
	CommonArgs
	OutputPath string `arg:"positional,required" help:"path to write generated Github Actions definition to"`
}

func (g GenerateArgs) CreatePipeline() (Pipeline, error) {
	return ParsePipeline(ActionArgs{
		CommonArgs: g.CommonArgs,
		Id:         "",
		CurrentSha: "",
		Force:      false,
	}, "")
}

type GenerateCommand struct {
	GenerateArgs
}

func (c GenerateCommand) Run() error {
	pipeline, err := c.CreatePipeline()
	if err != nil {
		return err
	}

	return pipeline.ToGitHubWorkflow().WriteYaml(c.OutputPath)
}

type BuildArtifactCommand struct {
	ActionArgs
}

func (c BuildArtifactCommand) Run() error {
	pipeline, err := c.CreatePipeline()
	if err != nil {
		return err
	}

	return pipeline.BuildArtifact(c.Id)
}

type DeployApplicationCommand struct {
	ActionArgs
}

func (c DeployApplicationCommand) Run() error {
	pipeline, err := c.CreatePipeline()
	if err != nil {
		return err
	}

	return pipeline.DeployApplication(c.Id)
}
