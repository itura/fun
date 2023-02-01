package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveSecrets(t *testing.T) {
	builder := NewTestBuilder("currentSha")
	app := PostgresHelmChart(builder)
	expectedSecretMappings := map[string]string{
		"postgresql_auth_password": "${{ steps.secrets-gcp-project.outputs.pg-password }}",
		"postgresql_auth_username": "${{ secrets.pg-username }}",
	}

	resolvedSecrets := app.ResolveSecrets()

	assert.Equal(t, expectedSecretMappings, resolvedSecrets)
}
