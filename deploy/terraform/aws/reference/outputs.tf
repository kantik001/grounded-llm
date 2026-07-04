output "alb_dns_name" {
  description = "Public ALB DNS name (point your domain here)"
  value       = aws_lb.main.dns_name
}

output "rds_endpoint" {
  description = "Postgres hostname (private — reachable from ECS tasks)"
  value       = aws_db_instance.postgres.address
}

output "ecs_cluster_name" {
  value = aws_ecs_cluster.main.name
}

output "database_url_hint" {
  description = "DATABASE_URL format for server (password redacted in state)"
  value       = "postgres://${var.db_username}:<password>@${aws_db_instance.postgres.address}:5432/${var.db_name}?sslmode=require"
  sensitive   = true
}
