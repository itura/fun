package build

import (
	"fmt"
)

func ValidPipelineConfig(builder TestBuilder) PipelineConfig {
	artifactApi := builder.Artifact("api", "packages/api")
	artifactClient := builder.Artifact("client", "packages/client")
	appInfra := TerraformConfig(builder, "infra", "tf/main")
	appDatabase := PostgresHelmChart(builder, appInfra)
	appWebsite := WebsiteHelmChart(builder, artifactClient, artifactApi, appInfra, appDatabase)
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
	)
}

func TerraformConfig(builder TestBuilder, id string, path string) Application {
	app := builder.Application(id, path, typeTerraform).
		AddStep(
			CheckoutRepoStep(),
			SetupGoStep(),
			GcpAuthStep("${{ secrets.WORKLOAD_IDENTITY_PROVIDER }}", "${{ secrets.BUILD_AGENT_SA }}"),
		)
	return app
}

func PostgresHelmChart(builder TestBuilder, upstreams ...Job) Application {
	return builder.Application("db", "helm/db", typeHelm, upstreams...).
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

func WebsiteHelmChart(builder TestBuilder, upstreams ...Job) Application {
	return builder.Application("website", "helm/website", typeHelm, upstreams...).
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
}

func NewTestBuilder(currentSha string) TestBuilder {
	return TestBuilder{
		currentSha: currentSha,
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

func (b TestBuilder) Application(id string, path string, appType ApplicationType, upstreams ...Job) Application {
	return Application{
		Id:                id,
		Path:              path,
		Repository:        b.repository(),
		KubernetesCluster: b.clusterConfig,
		CurrentSha:        b.currentSha,
		Namespace:         "",
		RuntimeArgs:       nil,
		Upstreams:         upstreams,
		Type:              appType,
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
