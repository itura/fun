terraform {
  required_version = ">=1.3"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">=4.44.1"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">=2.16.0"
    }
  }
}