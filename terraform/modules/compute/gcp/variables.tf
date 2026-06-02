############################################
# Uniform compute-module contract.
# Every cloud's compute module declares these exact inputs so the
# per-provider roots can wire them identically.
############################################

variable "server_name" {
  description = "Name/label for the server instance"
  type        = string
  default     = "db-router"
}

variable "region" {
  description = "Provider region/location/zone slug"
  type        = string
}

variable "instance_size" {
  description = "Provider instance size/type slug"
  type        = string
}

variable "image" {
  description = "OS image slug/id (Debian 12/13). May be ignored where the provider needs a structured image reference."
  type        = string
  default     = ""
}

variable "ssh_public_key" {
  description = "Public SSH key content to authorize on the server"
  type        = string
}

variable "allowed_ips" {
  description = "CIDRs allowed to reach SSH (port 22). HTTP/HTTPS stay open for Caddy/ACME."
  type        = list(string)
  default     = ["0.0.0.0/0", "::/0"]
}

variable "tags" {
  description = "Tags/labels to apply where the provider supports them"
  type        = list(string)
  default     = ["terraform", "db-router", "grpc"]
}
