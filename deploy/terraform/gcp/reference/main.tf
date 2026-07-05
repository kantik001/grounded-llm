# Grounded LLM — GCP reference (Cloud Run + Cloud SQL)
#
# Reference stack for teams on Google Cloud. Extend with VPC connector,
# Secret Manager, and GCS buckets for data/chroma before production.
#
# See docs/en/TERRAFORM.md

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_sql_database_instance" "postgres" {
  name             = "${var.name_prefix}-postgres"
  database_version = "POSTGRES_16"
  region           = var.region

  settings {
    tier = var.db_tier
    ip_configuration {
      ipv4_enabled = true
    }
  }

  deletion_protection = false
}

resource "google_sql_database" "grounded" {
  name     = var.db_name
  instance = google_sql_database_instance.postgres.name
}

resource "google_sql_user" "grounded" {
  name     = var.db_username
  instance = google_sql_database_instance.postgres.name
  password = var.db_password
}

resource "google_cloud_run_v2_service" "python" {
  name     = "${var.name_prefix}-python"
  location = var.region

  template {
    containers {
      image = var.python_image
      ports {
        container_port = 5000
      }
      env {
        name  = "PYTHON_SERVICE_PORT"
        value = "5000"
      }
      env {
        name  = "VECTOR_STORE"
        value = var.vector_store
      }
    }
    scaling {
      max_instance_count = 2
    }
  }
}

resource "google_cloud_run_v2_service" "server" {
  name     = "${var.name_prefix}-server"
  location = var.region

  template {
    containers {
      image = var.server_image
      ports {
        container_port = 8080
      }
      env {
        name  = "DATABASE_URL"
        value = "postgres://${var.db_username}:${var.db_password}@${google_sql_database_instance.postgres.public_ip_address}:5432/${var.db_name}?sslmode=require"
      }
      env {
        name  = "PYTHON_RAG_URL"
        value = "${google_cloud_run_v2_service.python.uri}/rag/context"
      }
    }
    scaling {
      max_instance_count = 2
    }
  }
}

resource "google_cloud_run_v2_service_iam_member" "server_public" {
  location = google_cloud_run_v2_service.server.location
  name     = google_cloud_run_v2_service.server.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
