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

# Resource for creating a random password. We are creating x many passwords,
# each y characters long.
resource "random_password" "store_service_password" {
  count            = length(var.json_config.service_configuration[*])
  length           = var.password_length
  override_special = "%^#;_"
}

# Create our database users looping through the service configuration provided
# in the config.json file.
resource "mongodbatlas_database_user" "store_service_user" {
  count              = length(var.json_config.service_configuration)
  username           = var.json_config.service_configuration[count.index].serviceName
  password           = random_password.store_service_password[count.index].result
  auth_database_name = "admin"
  project_id         = var.project_id

  dynamic "roles" {
    for_each = var.json_config.service_configuration[count.index].mongoCollection[*]
    content {
      role_name       = "read"
      database_name   = var.json_config.service_configuration[count.index].mongoDatabase
      collection_name = roles.value
    }
  }
}

# Generate a local file output (from template) that contains our connection
# strings, formatted as: mongodb+srv://[username]:[password]@[cluster]/[db]/[collection]
resource "local_file" "connection_strings" {
  count = length(var.json_config.service_configuration)
  content = templatefile("../../templates/connection_strings.json.tpl", {
    service_config    = var.json_config.service_configuration[count.index],
    connection_string = replace(var.connection_string, "/^mongodb\\+srv:\\/\\//", "")
    password          = random_password.store_service_password[count.index].result
  })

  filename = "${var.config_out_dir}/${var.env_id}-${var.json_config.service_configuration[count.index].serviceName}_connection_string.outputs.json"
}
