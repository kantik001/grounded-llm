# Grounded LLM — Azure reference (Container Apps + PostgreSQL)
#
# See docs/en/TERRAFORM.md

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "main" {
  name     = "${var.name_prefix}-rg"
  location = var.location
  tags     = var.tags
}

resource "azurerm_postgresql_flexible_server" "postgres" {
  name                   = "${var.name_prefix}-pg"
  resource_group_name    = azurerm_resource_group.main.name
  location               = azurerm_resource_group.main.location
  version                = "16"
  administrator_login    = var.db_username
  administrator_password = var.db_password
  storage_mb             = 32768
  sku_name               = var.db_sku
  zone                   = "1"
  tags                   = var.tags
}

resource "azurerm_postgresql_flexible_server_database" "grounded" {
  name      = var.db_name
  server_id = azurerm_postgresql_flexible_server.postgres.id
  charset   = "UTF8"
  collation = "en_US.utf8"
}

resource "azurerm_container_app_environment" "env" {
  name                = "${var.name_prefix}-cae"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tags                = var.tags
}

resource "azurerm_container_app" "python" {
  name                         = "${var.name_prefix}-python"
  container_app_environment_id = azurerm_container_app_environment.env.id
  resource_group_name          = azurerm_resource_group.main.name
  revision_mode                = "Single"
  tags                         = var.tags

  template {
    container {
      name   = "python"
      image  = var.python_image
      cpu    = 1.0
      memory = "2Gi"
      env {
        name  = "PYTHON_SERVICE_PORT"
        value = "5000"
      }
      env {
        name  = "VECTOR_STORE"
        value = var.vector_store
      }
    }
  }

  ingress {
    external_enabled = true
    target_port      = 5000
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }
}

resource "azurerm_container_app" "server" {
  name                         = "${var.name_prefix}-server"
  container_app_environment_id = azurerm_container_app_environment.env.id
  resource_group_name          = azurerm_resource_group.main.name
  revision_mode                = "Single"
  tags                         = var.tags

  template {
    container {
      name   = "server"
      image  = var.server_image
      cpu    = 0.5
      memory = "1Gi"
      env {
        name  = "DATABASE_URL"
        value = "postgres://${var.db_username}:${var.db_password}@${azurerm_postgresql_flexible_server.postgres.fqdn}:5432/${var.db_name}?sslmode=require"
      }
      env {
        name  = "PYTHON_RAG_URL"
        value = "https://${azurerm_container_app.python.ingress[0].fqdn}/rag/context"
      }
    }
  }

  ingress {
    external_enabled = true
    target_port      = 8080
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }
}
