package build

import (
	"fmt"
	"github.com/itura/fun/pkg/fun"
)

type Dependency struct {
	id         string
	path       string
	isArtifact bool
	upstreams  []string
}

func NewApplicationDependency(id, path string, upstreams ...string) Dependency {
	return Dependency{
		id:        id,
		path:      path,
		upstreams: upstreams,
	}
}

func NewArtifactDependency(id, path string) Dependency {
	return Dependency{
		id:         id,
		path:       path,
		isArtifact: true,
		upstreams:  nil,
	}
}

func (d Dependency) DependsOn(upstream ...string) Dependency {
	d.upstreams = append(d.upstreams, upstream...)
	return d
}

type Dependencies struct {
	deps fun.Config[Dependency]
}

func NewDependencies() Dependencies {
	return Dependencies{
		deps: map[string]Dependency{},
	}
}

func (d Dependencies) Set(id string, dep Dependency) Dependencies {
	d.deps[id] = dep
	return d
}

func ParseDependencies(config PipelineConfigRaw) Dependencies {
	d := Dependencies{deps: fun.NewConfig[Dependency]()}
	for _, artifact := range config.Artifacts {
		d.deps[artifact.Id] = NewArtifactDependency(artifact.Id, artifact.Path)
	}
	for _, application := range config.Applications {
		dep := NewApplicationDependency(application.Id, application.Path)
		for _, upstream := range application.Artifacts {
			dep = dep.DependsOn(upstream)
		}
		for _, upstream := range application.Dependencies {
			dep = dep.DependsOn(upstream)
		}
		d.deps[application.Id] = dep
	}
	// TODO add validations
	return d
}

func (d Dependencies) GetUpstreamJobIds(id string) []string {
	dep, ok := d.deps[id]
	if !ok {
		return nil
	}
	var results []string
	for _, upstream := range dep.upstreams {
		results = append(results, d.GetJobId(upstream))
	}
	return results
}

func (d Dependencies) GetJobId(id string) string {
	dep, ok := d.deps[id]
	if !ok {
		return ""
	}
	if dep.isArtifact {
		return fmt.Sprintf("build-%s", id)
	} else {
		return fmt.Sprintf("deploy-%s", id)
	}
}

func (d Dependencies) GetAllPaths(id string) []string {
	results := d.getAllPaths(id)
	return fun.RemoveDuplicate(results)
}

func (d Dependencies) getAllPaths(id string) []string {
	dep, ok := d.deps[id]
	if !ok {
		return nil
	}

	results := []string{dep.path}
	for _, upstreamId := range dep.upstreams {
		if _, ok := d.deps[upstreamId]; ok {
			results = append(results, d.getAllPaths(upstreamId)...)
		}
	}
	return results
}

type ChangeDetection interface {
	HasChanged(paths ...string) bool
}

type StaticChangeDetection struct {
	hasChanged bool
}

func NewAlwaysChanged() ChangeDetection {
	return StaticChangeDetection{hasChanged: true}
}

func NewNeverChanged() ChangeDetection {
	return StaticChangeDetection{hasChanged: false}
}

func (c StaticChangeDetection) HasChanged(...string) bool {
	return c.hasChanged
}

type GitChangeDetection struct {
	previousSha string
	paths       []string
	runner      CommandRunner
}

func NewGitChangeDetection(runner CommandRunner) (GitChangeDetection, error) {
	previousSha, err := runner.Output("git", "rev-list", "-n", "1", "HEAD~1")
	return GitChangeDetection{
		previousSha: previousSha,
		runner:      runner,
		paths: []string{
			".github/",
		},
	}, err
}

func (g GitChangeDetection) HasChanged(paths ...string) bool {
	hasChanged := false
	for _, path := range append(g.paths, paths...) {
		hasChanged = hasChanged || g.sourceHasChanged(path, g.previousSha)
	}
	return hasChanged
}

func (g GitChangeDetection) sourceHasChanged(path string, ref string) bool {
	err := g.runner.RunSilent("git", "diff", "--quiet", "HEAD", ref, "--", path)
	if err != nil {
		return true
	}
	return false
}
