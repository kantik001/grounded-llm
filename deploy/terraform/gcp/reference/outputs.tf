output "server_url" {
  description = "Cloud Run server URL"
  value       = google_cloud_run_v2_service.server.uri
}

output "python_url" {
  description = "Cloud Run Python RAG URL"
  value       = google_cloud_run_v2_service.python.uri
}

output "postgres_connection_name" {
  value = google_sql_database_instance.postgres.connection_name
}
