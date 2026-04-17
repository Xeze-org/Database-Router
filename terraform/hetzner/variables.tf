############################################
# Server
############################################

variable "server_name" {
  description = "Name of the Hetzner Cloud server"
  type        = string
  default     = "db-router"
}

variable "location" {
  description = "Hetzner Cloud location (fsn1=Falkenstein, nbg1=Nuremberg, hel1=Helsinki, ash=Ashburn, hil=Hillsboro)"
  type        = string
  default     = "fsn1"
}

variable "server_type" {
  description = "Server type (cx22=2vCPU/4GB, cx11=1vCPU/2GB, cx32=4vCPU/8GB)"
  type        = string
  default     = "cx22"
}

variable "image" {
  description = "OS image name"
  type        = string
  default     = "debian-12"
}

############################################
# SSH — hybrid key handling
#
# Container auto-generates a key → sets ssh_public_key via TF_VAR.
# Manual use with an existing Hetzner key → sets ssh_key_name instead.
############################################

variable "ssh_key_name" {
  description = "Name of an SSH key already uploaded to your Hetzner account (used when ssh_public_key is empty)"
  type        = string
  default     = ""
}

variable "ssh_public_key" {
  description = "Public key content to upload to Hetzner (set by the deploy container when auto-generating keys; leave empty for manual use)"
  type        = string
  default     = ""
}

############################################
# Domain / DNS
############################################

variable "domain" {
  description = "Base domain managed in Hetzner DNS"
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
  description = "CIDRs allowed to reach SSH and gRPC. Default is open — restrict in production."
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
