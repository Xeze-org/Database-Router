############################################
# Droplet
############################################

variable "droplet_name" {
  description = "Name of the DigitalOcean droplet"
  type        = string
  default     = "db-router"
}

variable "region" {
  description = "DigitalOcean region slug"
  type        = string
  default     = "blr1"
}

variable "droplet_size" {
  description = "Droplet size slug (1 GB is enough for light workloads)"
  type        = string
  default     = "s-1vcpu-2gb"
}

variable "image" {
  description = "OS image slug"
  type        = string
  default     = "debian-13-x64"
}

############################################
# SSH — hybrid key handling
#
# Container auto-generates a key → sets ssh_public_key via TF_VAR.
# Manual use with an existing DO key → sets ssh_key_name instead.
############################################

variable "ssh_key_name" {
  description = "Name of an SSH key already uploaded to your DO account (used when ssh_public_key is empty)"
  type        = string
  default     = "ayush"
}

variable "ssh_public_key" {
  description = "Public key content to upload to DO (set by the deploy container when auto-generating keys; leave empty for manual use)"
  type        = string
  default     = ""
}

############################################
# Domain / DNS
############################################

variable "domain" {
  description = "Base domain managed in DigitalOcean DNS"
  type        = string
  default     = "0.xeze.org"
}

variable "subdomain" {
  description = "Subdomain for the db-router A record (e.g. 'db' → db.0.xeze.org)"
  type        = string
  default     = "db"
}

############################################
# Database usernames (passwords are auto-generated)
############################################

variable "postgres_user" {
  description = "PostgreSQL admin username"
  type        = string
  default     = "admin"
}

variable "postgres_db" {
  description = "Default PostgreSQL database"
  type        = string
  default     = "unified_db"
}

variable "mongo_user" {
  description = "MongoDB root username"
  type        = string
  default     = "admin"
}

############################################
# Access control
############################################

variable "allowed_ips" {
  description = "CIDRs allowed to reach SSH and gRPC. Default is your IP only — never use 0.0.0.0/0 in production."
  type        = list(string)
  default     = ["0.0.0.0/0", "::/0"]
}

############################################
# Ports
############################################

variable "grpc_port" {
  description = "gRPC server port"
  type        = number
  default     = 50051
}

############################################
# mTLS
############################################

variable "enable_mtls" {
  description = "Enable mTLS on the gRPC server (Caddy verifies client certs at the edge)"
  type        = bool
  default     = true
}
