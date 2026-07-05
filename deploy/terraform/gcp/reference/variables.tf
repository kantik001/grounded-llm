variable "project_id" {
  type        = string
  description = "GCP project id"
}

variable "region" {
  type    = string
  default = "europe-west1"
}

variable "name_prefix" {
  type    = string
  default = "grounded-llm"
}

variable "db_tier" {
  type    = string
  default = "db-f1-micro"
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

variable "vector_store" {
  type    = string
  default = "chroma"
}
