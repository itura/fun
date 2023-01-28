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

func (g GCP) Project() string {
	return g.Config["project"]
}

func (g GCP) AuthStep() GitHubActionsStep {
	return GitHubActionsStep{
		Name: "Authenticate to GCloud via Service Account",
		Uses: "google-github-actions/auth@v1",
		With: map[string]interface{}{
			"workload_identity_provider": formatSecretValue(g.Config["workloadIdentityProvider"]),
			"service_account":            formatSecretValue(g.Config["serviceAccount"]),
		},
	}
}

func (g GCP) Validate(key string) ValidationErrors {
	errs := NewValidationErrors(key)
	if _, ok := g.Config["project"]; !ok {
		errs = errs.Put("project", CloudProviderMissingField(g.Type()))
	}
	if _, ok := g.Config["serviceAccount"]; !ok {
		errs = errs.Put("serviceAccount", CloudProviderMissingField(g.Type()))
	}
	if _, ok := g.Config["workloadIdentityProvider"]; !ok {
		errs = errs.Put("workloadIdentityProvider", CloudProviderMissingField(g.Type()))
	}
	return errs
}
