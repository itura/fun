package build

import (
	"fmt"
)

type TestBuilder struct {
	project         string
	currentSha      string
	clusterConfig   ClusterConfig
	artifactRepo    ArtifactRepository
	secretProviders SecretProviders
}

func NewTestBuilder(project, currentSha string) TestBuilder {
	return TestBuilder{
		project:    project,
		currentSha: currentSha,
		artifactRepo: ArtifactRepository{
			Host: "us-central1-docker.pkg.dev",
			Name: "repo-name",
		},
		secretProviders: SecretProviders{
			"princess-pup": {
				Type:   typeGcp,
				Config: map[string]string{"project": "princess-pup"},
			},
			"github": {
				Type:   typeGithub,
				Config: nil,
			},
		},
		clusterConfig: ClusterConfig{
			Name:     "cluster-name",
			Location: "uscentral1",
		},
	}
}

func (b TestBuilder) Artifact(id string, path string) Artifact {
	return Artifact{
		Id:              id,
		Path:            path,
		Project:         b.project,
		Repository:      b.repository(),
		Host:            b.artifactRepo.Host,
		CurrentSha:      b.currentSha,
		Type:            "app",
		hasDependencies: false,
		hasChanged:      true,
	}
}

func (b TestBuilder) repository() string {
	return fmt.Sprintf("%s/%s/%s", b.artifactRepo.Host, b.project, b.artifactRepo.Name)
}

func (b TestBuilder) Application(id string, path string, upstreams ...Job) Application {
	return Application{
		Id:                id,
		Path:              path,
		ProjectId:         b.project,
		Repository:        b.repository(),
		KubernetesCluster: b.clusterConfig,
		CurrentSha:        b.currentSha,
		Namespace:         "app-namespace",
		Values:            []HelmValue{},
		Upstreams:         upstreams,
		Type:              "helm",
		Secrets:           map[string][]HelmSecretValue{},
		SecretProviders:   b.secretProviders,
		hasDependencies:   false,
		hasChanged:        true,
	}
}

func getHelmApplication() Application {
	return Application{
		Id:         "db",
		Path:       "helm/db",
		ProjectId:  "projectId",
		Repository: "us-central1-docker.pkg.dev/projectId/repo-name",
		KubernetesCluster: ClusterConfig{
			Name:     "cluster-name",
			Location: "uscentral1",
		},
		CurrentSha: "currentSha",
		Namespace:  "app-namespace",
		Values: []HelmValue{
			{
				Key:   "postgresql.dbName",
				Value: "my-db",
			},
		},
		Upstreams: nil,
		Type:      ApplicationType("helm"),
		Secrets:   getValidSecrets(),
		SecretProviders: map[string]SecretProvider{
			"princess-pup": {
				Type:   SecretProviderType("gcp"),
				Config: map[string]string{"project": "princess-pup"},
			},
			"github": {
				Type:   SecretProviderType("github-actions"),
				Config: nil,
			},
		},
		hasDependencies: false,
		hasChanged:      true,
	}
}

func TestArgs(configPath string) ActionArgs {
	return ActionArgs{
		CommonArgs: CommonArgs{
			ConfigPath: configPath,
			Self:       false,
		},
		Id:         "test",
		CurrentSha: "currentSha",
		ProjectId:  "projectId",
		Force:      false,
	}
}

func getValidSecrets() map[string][]HelmSecretValue {
	return map[string][]HelmSecretValue{
		"princess-pup": {
			{
				HelmKey:    "postgresql.auth.password",
				SecretName: "pg-password",
			},
		},
		"github": {
			{
				HelmKey:    "postgresql.auth.username",
				SecretName: "pg-username",
			},
		},
	}
}

func getAppArtifact() Artifact {
	return Artifact{
		Id:              "api",
		Path:            "packages/api",
		Project:         "projectId",
		Repository:      "us-central1-docker.pkg.dev/projectId/repo-name",
		Host:            "us-central1-docker.pkg.dev",
		CurrentSha:      "currentSha",
		Type:            ArtifactType("app"),
		hasDependencies: false,
		hasChanged:      true,
	}
}
