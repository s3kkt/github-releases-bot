provider "google" {
  project = var.project_id
  region  = var.region
}

terraform {
  backend "gcs" {
    bucket = "gh-bot-tfstate"
    prefix = "terraform.tfstate"
  }
}