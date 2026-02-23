# ERP SPPG Infrastructure - Google Cloud Platform
# Terraform configuration for production deployment

terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

# Variables
variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "asia-southeast2"
}

variable "zone" {
  description = "GCP Zone"
  type        = string
  default     = "asia-southeast2-a"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "prod"
}

# Provider configuration
provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

# Enable required APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "compute.googleapis.com",
    "sql.googleapis.com",
    "storage.googleapis.com",
    "monitoring.googleapis.com",
    "logging.googleapis.com",
    "cloudresourcemanager.googleapis.com"
  ])
  
  project = var.project_id
  service = each.value
  
  disable_dependent_services = true
}

# VPC Network
resource "google_compute_network" "erp_network" {
  name                    = "erp-sppg-network"
  auto_create_subnetworks = false
  
  depends_on = [google_project_service.required_apis]
}

# Subnet
resource "google_compute_subnetwork" "erp_subnet" {
  name          = "erp-sppg-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.erp_network.id
  
  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.1.0.0/16"
  }
  
  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.2.0.0/16"
  }
}

# Cloud SQL Instance (PostgreSQL with HA)
resource "google_sql_database_instance" "erp_db" {
  name             = "erp-sppg-db-${var.environment}"
  database_version = "POSTGRES_15"
  region           = var.region
  
  settings {
    tier                        = "db-custom-2-4096"
    availability_type          = "REGIONAL"  # High Availability
    disk_type                  = "PD_SSD"
    disk_size                  = 100
    disk_autoresize           = true
    disk_autoresize_limit     = 500
    
    backup_configuration {
      enabled                        = true
      start_time                    = "03:00"
      location                      = var.region
      point_in_time_recovery_enabled = true
      transaction_log_retention_days = 7
      backup_retention_settings {
        retained_backups = 30
        retention_unit   = "COUNT"
      }
    }
    
    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.erp_network.id
      require_ssl     = true
    }
    
    database_flags {
      name  = "log_statement"
      value = "all"
    }
    
    database_flags {
      name  = "log_min_duration_statement"
      value = "1000"
    }
    
    maintenance_window {
      day          = 7  # Sunday
      hour         = 4  # 4 AM
      update_track = "stable"
    }
  }
  
  deletion_protection = true
  
  depends_on = [
    google_project_service.required_apis,
    google_service_networking_connection.private_vpc_connection
  ]
}

# Cloud SQL Database
resource "google_sql_database" "erp_database" {
  name     = "erp_sppg"
  instance = google_sql_database_instance.erp_db.name
}

# Cloud SQL User
resource "google_sql_user" "erp_user" {
  name     = "erp_sppg_user"
  instance = google_sql_database_instance.erp_db.name
  password = random_password.db_password.result
}

# Random password for database
resource "random_password" "db_password" {
  length  = 32
  special = true
}

# Private Service Connection for Cloud SQL
resource "google_compute_global_address" "private_ip_address" {
  name          = "private-ip-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.erp_network.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.erp_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}

# Cloud Storage Bucket for file uploads
resource "google_storage_bucket" "erp_storage" {
  name          = "erp-sppg-storage-${var.project_id}"
  location      = var.region
  force_destroy = false
  
  uniform_bucket_level_access = true
  
  versioning {
    enabled = true
  }
  
  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type = "Delete"
    }
  }
  
  cors {
    origin          = ["https://erp-sppg.example.com"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
}

# Cloud Storage Bucket for backups
resource "google_storage_bucket" "erp_backups" {
  name          = "erp-sppg-backups-${var.project_id}"
  location      = var.region
  force_destroy = false
  
  uniform_bucket_level_access = true
  
  versioning {
    enabled = true
  }
  
  lifecycle_rule {
    condition {
      age = 365
    }
    action {
      type = "Delete"
    }
  }
}

# Compute Engine Instance Template
resource "google_compute_instance_template" "erp_template" {
  name_prefix  = "erp-sppg-template-"
  machine_type = "e2-standard-2"
  
  disk {
    source_image = "cos-cloud/cos-stable"
    auto_delete  = true
    boot         = true
    disk_size_gb = 50
    disk_type    = "pd-ssd"
  }
  
  network_interface {
    network    = google_compute_network.erp_network.id
    subnetwork = google_compute_subnetwork.erp_subnet.id
    
    access_config {
      # Ephemeral public IP
    }
  }
  
  service_account {
    email = google_service_account.erp_service_account.email
    scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }
  
  metadata = {
    startup-script = file("${path.module}/startup-script.sh")
  }
  
  tags = ["erp-sppg", "http-server", "https-server"]
  
  lifecycle {
    create_before_destroy = true
  }
}

# Managed Instance Group
resource "google_compute_region_instance_group_manager" "erp_group" {
  name   = "erp-sppg-group"
  region = var.region
  
  base_instance_name = "erp-sppg"
  target_size        = 2
  
  version {
    instance_template = google_compute_instance_template.erp_template.id
  }
  
  named_port {
    name = "http"
    port = 80
  }
  
  named_port {
    name = "https"
    port = 443
  }
  
  auto_healing_policies {
    health_check      = google_compute_health_check.erp_health_check.id
    initial_delay_sec = 300
  }
}

# Health Check
resource "google_compute_health_check" "erp_health_check" {
  name               = "erp-sppg-health-check"
  check_interval_sec = 30
  timeout_sec        = 10
  
  http_health_check {
    port         = 80
    request_path = "/health"
  }
}

# Load Balancer - Backend Service
resource "google_compute_backend_service" "erp_backend" {
  name                  = "erp-sppg-backend"
  protocol              = "HTTP"
  timeout_sec           = 30
  enable_cdn           = true
  load_balancing_scheme = "EXTERNAL"
  
  backend {
    group           = google_compute_region_instance_group_manager.erp_group.instance_group
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }
  
  health_checks = [google_compute_health_check.erp_health_check.id]
  
  cdn_policy {
    cache_mode                   = "CACHE_ALL_STATIC"
    default_ttl                  = 3600
    max_ttl                      = 86400
    negative_caching             = true
    serve_while_stale            = 86400
  }
}

# Load Balancer - URL Map
resource "google_compute_url_map" "erp_url_map" {
  name            = "erp-sppg-url-map"
  default_service = google_compute_backend_service.erp_backend.id
  
  host_rule {
    hosts        = ["erp-sppg.example.com"]
    path_matcher = "allpaths"
  }
  
  path_matcher {
    name            = "allpaths"
    default_service = google_compute_backend_service.erp_backend.id
    
    path_rule {
      paths   = ["/api/*"]
      service = google_compute_backend_service.erp_backend.id
    }
  }
}

# Load Balancer - HTTPS Proxy
resource "google_compute_target_https_proxy" "erp_https_proxy" {
  name             = "erp-sppg-https-proxy"
  url_map          = google_compute_url_map.erp_url_map.id
  ssl_certificates = [google_compute_managed_ssl_certificate.erp_ssl_cert.id]
}

# SSL Certificate
resource "google_compute_managed_ssl_certificate" "erp_ssl_cert" {
  name = "erp-sppg-ssl-cert"
  
  managed {
    domains = ["erp-sppg.example.com"]
  }
}

# Global Forwarding Rule
resource "google_compute_global_forwarding_rule" "erp_forwarding_rule" {
  name       = "erp-sppg-forwarding-rule"
  target     = google_compute_target_https_proxy.erp_https_proxy.id
  port_range = "443"
}

# Service Account
resource "google_service_account" "erp_service_account" {
  account_id   = "erp-sppg-service"
  display_name = "ERP SPPG Service Account"
}

# IAM bindings for service account
resource "google_project_iam_member" "erp_storage_admin" {
  project = var.project_id
  role    = "roles/storage.admin"
  member  = "serviceAccount:${google_service_account.erp_service_account.email}"
}

resource "google_project_iam_member" "erp_sql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.erp_service_account.email}"
}

# Firewall Rules
resource "google_compute_firewall" "allow_http" {
  name    = "allow-http"
  network = google_compute_network.erp_network.name
  
  allow {
    protocol = "tcp"
    ports    = ["80"]
  }
  
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["http-server"]
}

resource "google_compute_firewall" "allow_https" {
  name    = "allow-https"
  network = google_compute_network.erp_network.name
  
  allow {
    protocol = "tcp"
    ports    = ["443"]
  }
  
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["https-server"]
}

# Outputs
output "database_connection_name" {
  value = google_sql_database_instance.erp_db.connection_name
}

output "database_private_ip" {
  value = google_sql_database_instance.erp_db.private_ip_address
}

output "storage_bucket_name" {
  value = google_storage_bucket.erp_storage.name
}

output "backup_bucket_name" {
  value = google_storage_bucket.erp_backups.name
}

output "load_balancer_ip" {
  value = google_compute_global_forwarding_rule.erp_forwarding_rule.ip_address
}