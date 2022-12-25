variable "project" {
  type = string
}

variable "service_account_id" {
  type = string
}

variable "k8s" {
  type    = bool
  default = true
}

variable "k8s_namespace" {
  type        = string
  default     = ""
  description = "required if k8s is true"
}

variable "workload_identity_namespace" {
  type    = string
  default = ""
  description = "required if k8s is true"
}


variable "users" {
  type    = list(string)
  default = []
}

variable "token_creators" {
  type    = list(string)
  default = []
}

locals {
  k8s_principal  = "serviceAccount:${var.workload_identity_namespace}[${var.k8s_namespace}/${var.service_account_id}]"
  k8s_principals = var.k8s ? [local.k8s_principal] : []
  users          = concat(var.users, local.k8s_principals)
}

resource "google_service_account" "sa" {
  project    = var.project
  account_id = var.service_account_id
}

resource "kubernetes_service_account" "sa" {
  count = var.k8s ? 1 : 0
  metadata {
    name        = google_service_account.sa.account_id
    namespace   = var.k8s_namespace
    annotations = {
      "iam.gke.io/gcp-service-account" = google_service_account.sa.email
    }
  }
}

resource "google_service_account_iam_binding" "users" {
  count              = length(local.users) > 0 ? 1 : 0
  service_account_id = google_service_account.sa.name
  role               = "roles/iam.workloadIdentityUser"
  members            = local.users
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
