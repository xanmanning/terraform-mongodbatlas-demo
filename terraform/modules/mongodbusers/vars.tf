# MongoDB Atlas project id
variable "project_id" {}

# Connection string provided by mongodbcluster module
variable "connection_string" {}

# Name for the cluster
variable "cluster_name" {}

# Length of passwords
variable "password_length" {}

# Environment being deployed
variable "env_id" {}

# Imported JSON configuration (see config.json in project root)
variable "json_config" {}

# Directory to write output
variable "config_out_dir" {}
