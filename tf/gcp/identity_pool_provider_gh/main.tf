variable "workload_identity_pool_id" {
  type = string
}

variable "name" {
  type = string
}

variable "repos" {
  type = list(string)
  default = []
}


resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_provider_id = "github"
  workload_identity_pool_id          = var.workload_identity_pool_id
  attribute_mapping = {
    "google.subject" = "assertion.sub"
    "attribute.actor" = "assertion.actor"
    "attribute.repository" = "assertion.repository"
  }
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

output "name" {
  value = google_iam_workload_identity_pool_provider.github.name
}

output "principals" {
  value = {for repo in var.repos: repo => "principalSet://iam.googleapis.com/${var.name}/attribute.repository/${repo}"}
}
