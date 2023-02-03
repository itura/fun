package build

type CloudProvider interface {
	AuthStep() GitHubActionsStep
	Validate(key string) ValidationErrors
}

type GCP struct {
	Config map[string]string
}

func (g GCP) Type() string {
	return "gcp"
}

func (g GCP) AuthStep() GitHubActionsStep {
	return GcpAuthStep(
		formatSecretValue(g.Config["workloadIdentityProvider"]),
		formatSecretValue(g.Config["serviceAccount"]),
	)
}

func (g GCP) Validate(key string) ValidationErrors {
	errs := NewValidationErrors(key)
	if _, ok := g.Config["serviceAccount"]; !ok {
		errs = errs.Put("serviceAccount", CloudProviderMissingField(g.Type()))
	}
	if _, ok := g.Config["workloadIdentityProvider"]; !ok {
		errs = errs.Put("workloadIdentityProvider", CloudProviderMissingField(g.Type()))
	}
	return errs
}
