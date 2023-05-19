output "database_address" {
  description = "The public IP address of the deployed instance"
  value       = google_sql_database_instance.instance.connection_name
}