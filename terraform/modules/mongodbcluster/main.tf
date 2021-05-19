# Define the minimum Terraform version and required version of the MongoDB
# Atlas provider for this environment.
terraform {
  required_version = ">= 0.15.0"
  required_providers {
    mongodbatlas = {
      source  = "mongodb/mongodbatlas"
      version = "0.9.1"
    }
  }
}

resource "mongodbatlas_cluster" "cluster" {
  project_id = var.project_id
  name       = var.cluster_name

  # M2 is 2, M5 is 5
  disk_size_gb = "2"

  provider_name               = "TENANT"
  backing_provider_name       = "AWS"
  provider_region_name        = "EU_WEST_1"
  provider_instance_size_name = "M2"

  mongo_db_major_version       = "4.4"
  auto_scaling_disk_gb_enabled = "false"
}
