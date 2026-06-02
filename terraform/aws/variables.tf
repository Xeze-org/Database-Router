############################################
# Compute (DigitalOcean defaults)
############################################

variable "server_name" {
  description = "Name of the server / droplet"
  type        = string
  default     = "db-router"
}

variable "region" {
  description = "AWS region (configures the provider)"
  type        = string
  default     = "us-east-1"
}

variable "instance_size" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.small"
}

variable "image" {
  description = "AMI id; empty = latest Debian 12 AMI (data source)"
  type        = string
  default     = ""
}

variable "tags" {
  description = "Tags applied to the droplet"
  type        = list(string)
  default     = ["terraform", "db-router", "grpc"]
}

############################################
# SSH
############################################

variable "ssh_public_key" {
  description = "Public SSH key uploaded to the server (set by the deployer; required for manual use)"
  type        = string
  default     = ""
}

############################################
# DNS (Cloudflare)
############################################

variable "domain" {
  description = "Base domain for the db-router record (e.g. 0.xeze.org)"
  type        = string
  default     = "0.xeze.org"
}

variable "subdomain" {
  description = "Subdomain for the A record (e.g. 'db' -> db.0.xeze.org)"
  type        = string
  default     = "db"
}

variable "cloudflare_zone" {
  description = "Cloudflare zone (registered domain) that manages the DNS"
  type        = string
  default     = "xeze.org"
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID. If empty, looked up by cloudflare_zone name."
  type        = string
  default     = ""
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
# Access control & ports
############################################

variable "allowed_ips" {
  description = "CIDRs allowed to reach SSH. Default is open — restrict in production."
  type        = list(string)
  default     = ["0.0.0.0/0", "::/0"]
}

variable "grpc_port" {
  description = "gRPC server port (internal; fronted by Caddy on 443)"
  type        = number
  default     = 50051
}
