variable "instance_type" {
  // This is necessary to work around a limitation in Terraform 0.12. If this is set to `map`, as intended, Terraform expects all values within the map to have the same type, which is not the case here.
  type        = string
  description = "A set of Docker Volumes to configure"
  default     = "f1-micro"
}

variable "instance_count" {
  description = "Number of instances to create."
  type        = number
  default     = 1
}

variable "project_id" {
  description = "The project ID to deploy resources into"
}

variable "subnetwork_project" {
  description = "The project ID where the desired subnetwork is provisioned"
}

variable "subnetwork" {
  type        = string
  default     = "gh-bot-subnetwork"
  description = "The name of the subnetwork to deploy instances into"
}

variable "instance_name" {
  description = "The desired name to assign to the deployed instance"
  default     = "gh-bot-vm"
}

variable "region" {
  description = "The GCP region to deploy instances into"
  type        = string
}

variable "zone" {
  description = "The GCP zone to deploy instances into"
  type        = string
}

variable "additional_metadata" {
  type        = map(string)
  description = "Additional metadata to attach to the instance"
  default     = {}
}

variable "client_email" {
  description = "Service account email address"
  type        = string
  default     = ""
  sensitive   = true
}

variable "cos_image_name" {
  description = "The forced COS image to use instead of latest"
  default     = "cos-stable-77-12371-89-0"
}

variable "bot_github_token" {
  description = "GitHub token"
  sensitive   = true
}

variable "bot_telegram_token" {
  description = "Telegram token"
  sensitive   = true
}

variable "bot_debug" {
  description = "Enable debug output for Bot"
}

variable "bot_update_interval" {
  description = "Repositories check interval"
}

variable "bot_database_user" {
  description = "Bot DB user"
}

variable "bot_database_pass" {
  description = "Bot DB password"
  sensitive   = true
}

variable "bot_database_host" {
  description = "Bot DB instance address"
}

variable "bot_database_port" {
  description = "Bot DB instance port"
}

variable "bot_database_name" {
  description = "Bot DB name"
}

#variable "database_address" {
#  description = "Bot DB name"
#}