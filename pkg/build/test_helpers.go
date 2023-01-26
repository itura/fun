package build

func getValidArtifacts() map[string]Artifact {
	return map[string]Artifact{
		"api": {
			Id:              "api",
			Path:            "packages/api",
			Project:         "projectId",
			Repository:      "us-central1-docker.pkg.dev/projectId/repo-name",
			Host:            "us-central1-docker.pkg.dev",
			CurrentSha:      "currentSha",
			Type:            ArtifactType("app"),
			hasDependencies: false,
			hasChanged:      true,
		},
	}
}

func getValidApplications() map[string]Application {
	return map[string]Application{
		"db": getHelmApplication(),
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
		Secrets: map[string][]HelmSecretValue{
			"princess-pup": {
				{
					HelmKey:    "postgresql.auth.password",
					SecretName: "pg-password",
				},
			},
		},
		SecretProviders: map[string]SecretProvider{
			"princess-pup": {
				Type:   SecretProviderType("gcp"),
				Config: map[string]string{"project": "princess-pup"},
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
