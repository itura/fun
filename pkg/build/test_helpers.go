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
		Secrets: []HelmSecretValue{
			{
				HelmKey:    "postgresql.auth.password",
				SecretName: "pg-password",
				Provider: SecretProvider{
					Type:   "github-actions",
					Config: nil,
				},
			},
		},
		hasDependencies: false,
		hasChanged:      true,
	}
}
