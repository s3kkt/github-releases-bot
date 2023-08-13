locals {
  #instance_name = format("%s-%s", var.instance_name, substr(md5(module.gce-container.container.image), 0, 8))
  config_path = "/etc/github-bot/config.yml"
}

resource "google_storage_bucket" "gcs_bucket" {
  name          = "gh-bot-tfstate"
  location      = "EU"
  force_destroy = false
}

resource "google_compute_network" "peering_network" {
  name                    = "private-network"
  auto_create_subnetworks = "false"
}

resource "google_compute_global_address" "private_ip_address" {
  name          = "private-ip-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.peering_network.id
}

resource "google_service_networking_connection" "default" {
  network                 = google_compute_network.peering_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}

resource "google_sql_database_instance" "instance" {
  region           = var.region
  database_version = "POSTGRES_14"

  depends_on = [google_service_networking_connection.default]

  settings {
    tier = "db-f1-micro"
    database_flags {
      name  = "cloudsql.iam_authentication"
      value = "on"
    }
    database_flags {
      name  = "max_connections"
      value = 100
    }
    ip_configuration {
      ipv4_enabled    = "true"
      private_network = google_compute_network.peering_network.id
      require_ssl     = false
    }
    disk_size = 10
    #    ip_configuration {
    #      authorized_networks {
    #        //value = google_compute_instance.vm.network_interface.0.access_config.0.nat_ip
    #        value = google_cloud_run_service.run.connection
    #      }
    #    }
  }
  deletion_protection = true
}

#resource "google_compute_network_peering_routes_config" "peering_routes" {
#  peering              = google_service_networking_connection.default.peering
#  network              = google_compute_network.peering_network.name
#  import_custom_routes = true
#  export_custom_routes = true
#}

#resource "google_sql_database" "postgres" {
#  name     = var.bot_database_name
#  instance = google_sql_database_instance.instance.name
#  #deletion_policy  = "ABANDON"
#}

resource "google_sql_user" "database-user" {
  name            = var.bot_database_user
  instance        = google_sql_database_instance.instance.name
  password        = var.bot_database_pass
  deletion_policy = "ABANDON"
}

#resource "google_cloud_run_service_iam_member" "member" {
#  service = google_cloud_run_service.run.name
#  location = google_cloud_run_service.run.location
#  role = "roles/run.invoker"
#  member = "allUsers"
#}

resource "google_cloud_run_service" "run" {
  name     = "gh-bot"
  location = "us-central1"
  template {
    spec {
      containers {
        image = var.docker_image
        ports {
          container_port = 8080
        }
        liveness_probe {
          http_get {
            path = "/live"
          }
        }
        env {
          name  = "BOT_GITHUB_TOKEN"
          value = var.bot_github_token
        }
        env {
          name  = "BOT_TELEGRAM_TOKEN"
          value = var.bot_telegram_token
        }
        env {
          name  = "BOT_DEBUG"
          value = false
        }
        env {
          name  = "BOT_UPDATE_INTERVAL"
          value = "10m"
        }
        env {
          name  = "BOT_DB_USER"
          value = var.bot_database_user
        }
        env {
          name  = "BOT_DB_PASS"
          value = var.bot_database_pass
        }
        env {
          name  = "BOT_DB_HOST"
          value = google_sql_database_instance.instance.connection_name
        }
        env {
          name  = "BOT_DB_PORT"
          value = 5432
        }
        env {
          name  = "BOT_DB_NAME"
          value = var.bot_database_name
        }
        args = ["-cloud=true"]
      }
      service_account_name = var.client_email
    }
    metadata {
      annotations = {
        "run.googleapis.com/cloudsql-instances" = google_sql_database_instance.instance.connection_name
      }
    }
  }
  autogenerate_revision_name = true
}