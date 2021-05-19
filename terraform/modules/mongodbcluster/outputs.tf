output "connection_string" {
  value = mongodbatlas_cluster.cluster.connection_strings[0].standard_srv
}
