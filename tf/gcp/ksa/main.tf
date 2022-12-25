

variable "account_id" {
  type = string
}

variable "email" {
  type = string
}

variable "namespace" {
  type = string
}

variable "workload_identity_namespace" {
  type = string
  default = null
}

resource "kubernetes_service_account" "sa" {
  metadata {
    name        = var.account_id
    namespace   = var.namespace
    annotations = {
      "iam.gke.io/gcp-service-account" = var.email
    }
  }
}

output "principal" {
  value = var.workload_identity_namespace ? "serviceAccount:${var.workload_identity_namespace}[${var.namespace}/${var.account_id}]" : ""
}