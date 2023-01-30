package build

import (
	"fmt"
)

func PostgresHelmChart(builder TestBuilder) Application {
	return builder.Application("db", "helm/db").
		AddValue("postgresql.dbName", "my-db").
		SetSecret("postgresql.auth.password", "princess-pup", "pg-password").
		SetSecret("postgresql.auth.username", "github", "pg-username")
}

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
				Type:   secretProviderTypeGcp,
				Config: map[string]string{"project": "princess-pup"},
			},
			"github": {
				Type:   secretProviderTypeGithub,
				Config: nil,
			},
		},
		clusterConfig: ClusterConfig{
			Name:     "cluster-name",
			Location: "uscentral1",
		},
	}
}

func (b TestBuilder) Artifact(id string, path string, upstreams ...Job) Artifact {
	return Artifact{
		Id:              id,
		Path:            path,
		Repository:      b.repository(),
		Host:            b.artifactRepo.Host,
		CurrentSha:      b.currentSha,
		Type:            "app",
		hasDependencies: len(upstreams) > 0,
		hasChanged:      true,
		Upstreams:       upstreams,
		CloudProvider:   b.cloudProvider(),
	}
}

func (b TestBuilder) repository() string {
	return fmt.Sprintf("%s/%s/%s", b.artifactRepo.Host, b.project, b.artifactRepo.Name)
}

func (b TestBuilder) cloudProvider() CloudProviderConfig {
	return CloudProviderConfig{
		Type: cloudProviderTypeGcp,
		Config: map[string]string{
			"project":                  b.project,
			"workloadIdentityProvider": "WORKLOAD_IDENTITY_PROVIDER",
			"serviceAccount":           "BUILD_AGENT_SA",
		},
	}
}

func (b TestBuilder) Application(id string, path string, upstreams ...Job) Application {
	return Application{
		Id:                id,
		Path:              path,
		Repository:        b.repository(),
		KubernetesCluster: b.clusterConfig,
		CurrentSha:        b.currentSha,
		Namespace:         "app-namespace",
		Values:            []HelmValue{},
		Upstreams:         upstreams,
		Type:              "helm",
		Secrets:           map[string][]HelmSecretValue{},
		SecretProviders:   b.secretProviders,
		hasDependencies:   len(upstreams) > 0,
		hasChanged:        true,
		CloudProvider:     b.cloudProvider(),
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
		Force:      false,
	}
}
