locals {
  instance_name = format("%s-%s", var.instance_name, substr(md5(module.gce-container.container.image), 0, 8))
  config_path   = "/etc/github-bot/config.yml"
}

resource "google_storage_bucket" "gcs_bucket" {
  name          = "gh-bot-tfstate"
  location      = "EU"
  force_destroy = false
}

resource "google_sql_database_instance" "instance" {
  #name             = var.bot_database_name
  region           = var.region
  database_version = "POSTGRES_14"
  settings {
    tier      = "db-f1-micro"
    disk_size = 10
    ip_configuration {
      authorized_networks {
        value = google_compute_instance.vm.network_interface.0.access_config.0.nat_ip
      }
    }
  }
  deletion_protection = "true"
}

resource "google_sql_user" "bot_role" {
  name     = var.bot_database_user
  password = var.bot_database_pass
  instance = google_sql_database_instance.instance.name
  type     = "CLOUD_IAM_USER"
}

resource "google_sql_database" "postgres" {
  name     = var.bot_database_name
  instance = google_sql_database_instance.instance.name
  #deletion_policy = "DELETE"
}

#module "gce-advanced-container" {
#  source = "../../"
#
#  container = {
#    image = "busybox"
#    command = [
#      "tail"
#    ]
#    args = [
#      "-f",
#      "/dev/null"
#    ]
#    securityContext = {
#      privileged : true
#    }
#    tty : true
#    env = [
#      {
#        name  = "EXAMPLE"
#        value = "VAR"
#      }
#    ]
#  }
#
#  restart_policy = "OnFailure"
#}

module "gce-container" {
  source  = "terraform-google-modules/container-vm/google"
  version = "3.1.0"

  cos_image_name = var.cos_image_name

  container = {
    image = "s3kkt/github-releases-bot:latest"

    volumeMounts = [
      {
        mountPath = local.config_path
        name      = "config"
        readOnly  = true
      },
    ]
  }

  volumes = [
    {
      name = "config"
      hostPath = {
        path = local.config_path
      }
    },
  ]

  restart_policy = "Always"
}

resource "google_compute_instance" "vm" {
  name           = "gh-bot"
  description    = "coreDNS with containers on CoS."
  tags           = ["gh-bot-node"]
  machine_type   = var.instance_type
  project        = var.project_id
  zone           = var.zone
  can_ip_forward = false

  labels = {
    container-vm = module.gce-container.vm_container_label
  }

  boot_disk {
    initialize_params {
      image = module.gce-container.source_image
    }
  }

  network_interface {
    subnetwork_project = var.subnetwork_project
    subnetwork         = "https://www.googleapis.com/compute/v1/projects/${var.project_id}/regions/${var.region}/subnetworks/default"
    access_config {}
  }

  metadata = merge(
    {
      gce-container-declaration = module.gce-container.metadata_value
      google-logging-enabled    = "true"
      google-monitoring-enabled = "false"
      #ssh-keys                  = "root:${file("~/.ssh/gcloud.pub")}"
    },
    var.additional_metadata,
  )

  service_account {
    email = var.client_email
    scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]
  }

  metadata_startup_script = templatefile(
    "${path.module}/config.yml.tftpl",
    {
      config_path         = local.config_path,
      bot_github_token    = var.bot_github_token,
      bot_telegram_token  = var.bot_telegram_token,
      bot_debug           = var.bot_debug,
      bot_update_interval = var.bot_update_interval,
      bot_database_user   = var.bot_database_user,
      bot_database_pass   = var.bot_database_pass,
      bot_database_host   = "var.database_address",
      bot_database_port   = var.bot_database_port,
      bot_database_name   = var.bot_database_name,
    }
  )

  #  metadata_startup_script = data.template_file.config.rendered
}

#data "template_file" "config" {
#  template = "${path.module}/config.yml.tftpl"
#  vars = {
#          config_path = local.config_path,
#          bot_github_token    = var.bot_github_token,
#          bot_telegram_token  = var.bot_telegram_token,
#          bot_debug           = var.bot_debug,
#          bot_update_interval = var.bot_update_interval,
#          bot_database_user   = var.bot_database_user,
#          bot_database_pass   = var.bot_database_pass,
#          bot_database_host   = "var.database_address",
#          bot_database_port   = var.bot_database_port,
#          bot_database_name   = var.bot_database_name,
#  }
#}