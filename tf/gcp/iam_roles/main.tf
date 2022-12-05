variable "target" {
  type = string
}

variable "type" {
  type = string
}

variable "principal" {
  type = string
}

variable "project" {
  type    = string
  default = null
}

variable "roles" {
  type = list(string)
}

locals {
  members = {
    "bucket"            = var.type == "bucket" ? var.roles : []
    "project"           = var.type == "project" ? var.roles : []
    "artifact_registry" = var.type == "artifact_registry" ? var.roles : []
  }
}

resource "google_project_iam_member" "project_iam_member" {
  for_each = toset(local.members["project"])
  project  = var.target
  member   = var.principal
  role     = each.key
}

resource "google_storage_bucket_iam_member" "bucket_iam_member" {
  for_each = toset(local.members["bucket"])
  bucket   = var.target
  member   = var.principal
  role     = each.key
}

resource "google_artifact_registry_repository_iam_member" "tf" {
  for_each   = toset(local.members["artifact_registry"])
  project    = var.project
  repository = var.target
  member     = var.principal
  role       = each.key
}