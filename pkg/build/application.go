package build

import (
	"fmt"
)

type HelmValue struct {
	Key   string
	Value string
}

type HelmSecretValue struct {
	HelmKey    string
	SecretName string
	Provider   SecretProvider
}

type Application struct {
	Id                string
	Path              string
	ProjectId         string
	Repository        string
	KubernetesCluster ClusterConfig
	CurrentSha        string
	Namespace         string
	Values            []HelmValue
	Upstreams         []Job
	Type              ApplicationType
	Secrets           []HelmSecretValue
	hasDependencies   bool
	hasChanged        bool
}

func (a Application) PrepareBuild() (Build, error) {
	switch a.Type {
	case typeHelm:
		return NewHelm(a), nil
	case typeTerraform:
		return NewTerraform(a), nil
	default:
		return NullBuild{}, fmt.Errorf("invalid application type %s", a.Type)
	}
}

func (a Application) JobId() string {
	return fmt.Sprintf("deploy-%s", a.Id)
}

func (a Application) HasDependencies() bool {
	return a.hasDependencies
}

func (a Application) Setup() string {

	if a.Type == typeHelm {

		return fmt.Sprintf(`
    - uses: google-github-actions/get-gke-credentials@v1
      with:
        cluster_name: %s
        location: %s
    - uses: azure/setup-helm@v3
      with:
        version: v3.10.2
`, a.KubernetesCluster.Name, a.KubernetesCluster.Location)
	} else if a.Type == typeTerraform {
		return `
    - uses: 'hashicorp/setup-terraform@v2'
      with:
        terraform_version: '1.3.6'
`
	} else {
		return ""
	}
}
