variable "project" {
  type = string
}

variable "service_account_id" {
  type = string
}

variable "users" {
  type    = list(string)
  default = []
}

variable "token_creators" {
  type    = list(string)
  default = []
}

resource "google_service_account" "sa" {
  project    = var.project
  account_id = var.service_account_id
}

resource "google_service_account_iam_binding" "users" {
  count              = length(var.users) > 0 ? 1 : 0
  service_account_id = google_service_account.sa.name
  role               = "roles/iam.workloadIdentityUser"
  members            = var.users
}

resource "google_service_account_iam_binding" "token_creators" {
  count              = length(var.token_creators) > 0 ? 1 : 0
  service_account_id = google_service_account.sa.name
  role               = "roles/iam.serviceAccountTokenCreator"
  members            = var.token_creators
}

output "account_id" {
  value = google_service_account.sa.account_id
}

output "email" {
  value = google_service_account.sa.email
}

output "principal" {
  value = "serviceAccount:${google_service_account.sa.email}"
}
