# Output a connection string for use elsewhere
output "connection_string" {
  value = mongodbatlas_cluster.cluster.connection_strings[0].standard_srv
}
