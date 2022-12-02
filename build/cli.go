package build

type PipelineCommand interface {
	Run() error
}

type CommonArgs struct {
	ConfigPath string `arg:"--config" default:"pipeline.yaml" help:"path to pipeline definition yaml"`
}

type ActionArgs struct {
	CommonArgs
	Id         string `arg:"positional,required"`
	CurrentSha string `arg:"--current-sha,required" help:"current git sha, used for change detection"`
	ProjectId  string `arg:"--project-id,required" help:"GCP project id"`
}

func (a ActionArgs) CreatePipeline() (Pipeline, error) {
	return NewPipeline(a.ConfigPath, a.ProjectId, a.CurrentSha, previousCommit())
}

type GenerateArgs struct {
	CommonArgs
	OutputPath string `arg:"positional,required" help:"path to write generated Github Actions definition to"`
}

func (g GenerateArgs) CreatePipeline() (Pipeline, error) {
	return NewPipeline(g.ConfigPath, "", "", "")
}

type GenerateCommand struct {
	GenerateArgs
}

func (c GenerateCommand) Run() error {
	pipeline, err := c.CreatePipeline()
	if err != nil {
		return err
	}

	return pipeline.ToGithubActions(c.OutputPath)
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
