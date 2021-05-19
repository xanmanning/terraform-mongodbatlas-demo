variable "json_config" {
  type        = string
  description = "Configuration file to read"
}

variable "env_id" {
  type        = string
  description = "Environemnt variable key for JSON config"
}

variable "project_id" {
  type        = string
  description = "MongoDB Atlas Project ID"

  validation {
    condition     = length(var.project_id) > 0
    error_message = "The Project ID needs to be defined."
  }
}
