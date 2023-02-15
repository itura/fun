package build

import (
	"testing"
)

func TestGitHubResolveSecrets(t *testing.T) {
	//builder := NewTestBuilder()
	//gitHubProvider := builder.secretProviders["github"].Impl("github")
	//gitHubProvider = gitHubProvider.Add("postgresql.auth.password", "pg-password")
	//gitHubProvider = gitHubProvider.Add("postgresql.auth.username", "pg-username")
	//
	//expectedSecretMappings := map[string]string{
	//	"postgresql_auth_password": "${{ secrets.pg-password }}",
	//	"postgresql_auth_username": "${{ secrets.pg-username }}",
	//}
	//
	//runtimeArgs := gitHubProvider.GenerateEnvMap()
	//
	//assert.Equal(t, expectedSecretMappings, runtimeArgs)
}

func TestGcpResolveSecrets(t *testing.T) {
	//builder := NewTestBuilder()
	//gcpProvider := builder.secretProviders["gcp-project"].Impl("gcp-project")
	//gcpProvider = gcpProvider.Add("postgresql.auth.password", "pg-password")
	//gcpProvider = gcpProvider.Add("postgresql.auth.username", "pg-username")
	//
	//expectedSecretMappings := map[string]string{
	//	"postgresql_auth_password": "${{ steps.secrets-gcp-project.outputs.pg-password }}",
	//	"postgresql_auth_username": "${{ steps.secrets-gcp-project.outputs.pg-username }}",
	//}
	//
	//runtimeArgs := gcpProvider.GenerateEnvMap()
	//
	//assert.Equal(t, expectedSecretMappings, runtimeArgs)
}
