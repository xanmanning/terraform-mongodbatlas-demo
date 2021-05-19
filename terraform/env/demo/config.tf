locals {
  # Read JSON config into a local variable
  json_config = jsondecode(file(var.json_config))[var.env_id]
}
