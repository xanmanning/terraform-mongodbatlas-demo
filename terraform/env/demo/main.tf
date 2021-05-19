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

# We're going to use environment variables for this code, but something can
# be done with:
#  - public_key = "atlas_public_api_key"
#  - private_key  = "atlas_private_api_key"
provider "mongodbatlas" {}

# Import our MongoDB Cluster module
module "mongodbcluster" {
  source = "../../modules/mongodbcluster"

  cluster_name = local.json_config.mongoCluster
  project_id   = var.project_id
}

module "mongodbusers" {
  source = "../../modules/mongodbusers"

  cluster_name      = local.json_config.mongoCluster
  json_config       = local.json_config
  project_id        = var.project_id
  env_id            = var.env_id
  connection_string = module.mongodbcluster.connection_string
  config_out_dir    = "../../../outputs"
}
