variable "name_prefix" {
  type    = string
  default = "grounded-llm"
}

variable "location" {
  type    = string
  default = "westeurope"
}

variable "db_sku" {
  type    = string
  default = "B_Standard_B1ms"
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

variable "tags" {
  type    = map(string)
  default = { project = "grounded-llm", managed_by = "terraform" }
}
