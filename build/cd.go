package build

type ChangeDetection interface {
	HasChanged() bool
}

type GitChangeDetection struct {
	previousSha string
	paths       []string
}

func NewGitChangeDetection(previousSha string) GitChangeDetection {
	return GitChangeDetection{
		previousSha: previousSha,
		paths: []string{
			".github/",
		},
	}
}

func (g GitChangeDetection) AddPaths(paths ...string) GitChangeDetection {
	g.paths = append(g.paths, paths...)
	return g
}

func (g GitChangeDetection) HasChanged() bool {
	hasChanged := false
	for _, path := range g.paths {
		hasChanged = hasChanged || g.sourceHasChanged(path, g.previousSha)
	}
	return hasChanged
}

func (g GitChangeDetection) sourceHasChanged(path string, ref string) bool {
	err := commandSilent("git", "diff", "--quiet", "HEAD", ref, "--", path)
	if err != nil {
		return true
	}
	return false
}
