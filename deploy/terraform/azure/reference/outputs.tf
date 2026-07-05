output "server_fqdn" {
  description = "Container App server FQDN"
  value       = azurerm_container_app.server.ingress[0].fqdn
}

output "python_fqdn" {
  description = "Container App Python RAG FQDN"
  value       = azurerm_container_app.python.ingress[0].fqdn
}

output "postgres_fqdn" {
  value = azurerm_postgresql_flexible_server.postgres.fqdn
}
