package build

type InvalidSecretProvider struct{}

func (m *InvalidSecretProvider) Error() string {
	return "Invalid secret provider reference"
}
