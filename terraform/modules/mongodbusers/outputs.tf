output "password" {
  value     = random_password.store_service_password
  sensitive = true
}
