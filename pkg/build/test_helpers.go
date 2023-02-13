package build

import (
	"fmt"
)

func ValidPipelineConfig(builder TestBuilder) PipelineConfig {
	artifactApi := builder.Artifact("api", "packages/api")
	artifactClient := builder.Artifact("client", "packages/client")
	appInfra := TerraformConfig(builder, "infra", "tf/main")
	appDatabase := PostgresHelmChart(builder, appInfra.Id)
	appWebsite := WebsiteHelmChart(builder, artifactClient.Id, artifactApi.Id, appInfra.Id, appDatabase.Id)
	return SuccessfulParse(
		"My Build",
		map[string]Artifact{
			"api":    artifactApi,
			"client": artifactClient,
		},
		map[string]Application{
			"infra":   appInfra,
			"db":      appDatabase,
			"website": appWebsite,
		},
		builder.deps,
	)
}

func TerraformConfig(builder TestBuilder, id string, path string) Application {
	app := builder.Application(id, path, applicationTypeTerraform).
		AddStep(
			CheckoutRepoStep(),
			SetupGoStep(),
			GcpAuthStep("${{ secrets.WORKLOAD_IDENTITY_PROVIDER }}", "${{ secrets.BUILD_AGENT_SA }}"),
		)
	return app
}

func PostgresHelmChart(builder TestBuilder, upstreams ...string) Application {
	return builder.Application("db", "helm/db", applicationTypeHelm, upstreams...).
		SetNamespace("db-namespace").
		AddRuntimeArg("postgresql.dbName", "my-db").
		AddRuntimeArg("postgresql.auth.password", "${{ steps.secrets-gcp-project.outputs.pg-password }}").
		AddRuntimeArg("postgresql.auth.username", "${{ secrets.pg-username }}").
		AddStep(
			CheckoutRepoStep(),
			SetupGoStep(),
			GcpAuthStep("${{ secrets.WORKLOAD_IDENTITY_PROVIDER }}", "${{ secrets.BUILD_AGENT_SA }}"),
			FetchGcpSecretsStep("gcp-project", "gcp-project", "pg-password"),
		)
}

func WebsiteHelmChart(builder TestBuilder, upstreams ...string) Application {
	return builder.Application("website", "helm/website", applicationTypeHelm, upstreams...).
		SetNamespace("website-namespace").
		AddRuntimeArg("app-name", "website").
		AddRuntimeArg("client.secrets.clientId", "${{ steps.secrets-gcp-project.outputs.client-id }}").
		AddRuntimeArg("client.secrets.clientSecret", "${{ steps.secrets-gcp-project.outputs.client-secret }}").
		AddRuntimeArg("client.secrets.nextAuthSecret", "${{ steps.secrets-gcp-project.outputs.next-auth-secret }}").
		AddRuntimeArg("client.secrets.nextAuthUrl", "${{ steps.secrets-gcp-project.outputs.next-auth-url }}").
		AddStep(
			CheckoutRepoStep(),
			SetupGoStep(),
			GcpAuthStep("${{ secrets.WORKLOAD_IDENTITY_PROVIDER }}", "${{ secrets.BUILD_AGENT_SA }}"),
			FetchGcpSecretsStep("gcp-project", "gcp-project", "client-id", "client-secret", "next-auth-url", "next-auth-secret"),
		)
}

type TestBuilder struct {
	project         string
	currentSha      string
	clusterConfig   ClusterConfig
	artifactRepo    ArtifactRepository
	secretProviders SecretProviders
	deps            Dependencies
}

func NewTestBuilder() TestBuilder {
	return TestBuilder{
		currentSha: "currentSha",
		artifactRepo: ArtifactRepository{
			Host: "us-central1-docker.pkg.dev",
			Name: "gcp-project/repo-name",
		},
		secretProviders: SecretProviders{
			"gcp-project": {
				Type:   secretProviderTypeGcp,
				Config: map[string]string{"project": "gcp-project"},
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
		deps: NewDependencies(),
	}
}

func (b TestBuilder) Artifact(id string, path string) Artifact {
	b.deps = b.deps.Set(id, NewArtifactDependency(id, path))
	return Artifact{
		Id:            id,
		Path:          path,
		Repository:    b.repository(),
		Host:          b.artifactRepo.Host,
		CurrentSha:    b.currentSha,
		hasChanged:    true,
		CloudProvider: b.cloudProvider(),
	}
}

func (b TestBuilder) repository() string {
	return fmt.Sprintf("%s/%s", b.artifactRepo.Host, b.artifactRepo.Name)
}

func (b TestBuilder) cloudProvider() CloudProviderConfig {
	return CloudProviderConfig{
		Type: cloudProviderTypeGcp,
		Config: map[string]string{
			"workloadIdentityProvider": "WORKLOAD_IDENTITY_PROVIDER",
			"serviceAccount":           "BUILD_AGENT_SA",
		},
	}
}

func (b TestBuilder) Application(id string, path string, appType ApplicationType, upstreams ...string) Application {
	b.deps = b.deps.Set(id, NewApplicationDependency(id, path, upstreams...))
	return Application{
		Id:                id,
		Path:              path,
		Repository:        b.repository(),
		KubernetesCluster: b.clusterConfig,
		CurrentSha:        b.currentSha,
		Namespace:         "",
		RuntimeArgs:       nil,
		Type:              appType,
		hasChanged:        true,
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
