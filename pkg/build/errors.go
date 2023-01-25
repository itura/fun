package build

import "fmt"

type MissingSecretProvider struct{}

func (m MissingSecretProvider) Error() string {
	return "Invalid secret provider reference"
}

type InvalidSecretProviderType struct {
	GivenType string
}

func (m InvalidSecretProviderType) Error() string {
	return fmt.Sprintf("Invalid provider type `%s` given", m.GivenType)
}
