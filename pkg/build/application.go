package build

import (
	"bytes"
	"fmt"
	"text/template"
)

type HelmValue struct {
	Key         string
	Value       string
	SecretValue string `yaml:"secretValue"`
	EnvValue    string `yaml:"envValue"`
}

type Application struct {
	Id         string
	Path       string
	ProjectId  string
	CurrentSha string
	Namespace  string
	Values     []HelmValue
	Upstreams  []Job
	Type       ApplicationType

	hasDependencies bool
	hasChanged      bool
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
	var err error
	tpl := template.New("workflow").
		Funcs(template.FuncMap{
			"secret":       formatSecretValue,
			"env":          formatEnvValue,
			"resolveValue": resolveValue,
			"resolveKey":   resolveKey,
		})
	if a.Type == typeHelm {
		tpl, err = tpl.Parse(`
    - uses: google-github-actions/get-gke-credentials@v1
      with:
        cluster_name: {{ secret "GKE_CLUSTER" }}
        location: {{ secret "GKE_LOCATION" }}
    - uses: azure/setup-helm@v3
      with:
        version: v3.10.2
`)
		if err != nil {
			panic(err)
		}

		var result bytes.Buffer
		if err = tpl.Execute(&result, a); err != nil {
			panic(err)
		}
		return result.String()
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
