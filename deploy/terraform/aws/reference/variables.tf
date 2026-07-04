variable "name_prefix" {
  type        = string
  description = "Resource name prefix"
  default     = "grounded-llm"
}

variable "aws_region" {
  type        = string
  description = "AWS region"
  default     = "eu-west-1"
}

variable "vpc_cidr" {
  type    = string
  default = "10.20.0.0/16"
}

variable "allowed_cidr_blocks" {
  type        = list(string)
  description = "CIDR blocks allowed to reach the ALB"
  default     = ["0.0.0.0/0"]
}

variable "db_instance_class" {
  type    = string
  default = "db.t4g.micro"
}

variable "db_username" {
  type    = string
  default = "grounded"
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "db_name" {
  type    = string
  default = "grounded"
}

variable "server_image" {
  type    = string
  default = "ghcr.io/kantik001/grounded-llm-server:latest"
}

variable "python_image" {
  type    = string
  default = "ghcr.io/kantik001/grounded-llm-python:latest"
}

variable "webapp_image" {
  type    = string
  default = "ghcr.io/kantik001/grounded-llm-webapp:latest"
}

variable "llm_base_url" {
  type    = string
  default = "https://openrouter.ai/api"
}

variable "llm_model" {
  type    = string
  default = "openrouter/free"
}

variable "llm_api_key_secret_arn" {
  type        = string
  description = "Secrets Manager ARN for LLM_API_KEY"
}

variable "admin_secret_arn" {
  type        = string
  description = "Secrets Manager ARN for ADMIN_SECRET"
}

variable "rag_service_token_secret_arn" {
  type        = string
  description = "Secrets Manager ARN for RAG_SERVICE_TOKEN"
}

variable "vector_store" {
  type        = string
  description = "VECTOR_STORE env for Python RAG (chroma or qdrant)"
  default     = "chroma"
}

variable "tags" {
  type    = map(string)
  default = { project = "grounded-llm", managed_by = "terraform" }
}
