variable "domain" {
  description = "Base domain for the db-router record (e.g. 0.xeze.org)"
  type        = string
}

variable "subdomain" {
  description = "Subdomain for the A record (e.g. 'db' -> db.<domain>)"
  type        = string
}

variable "cloudflare_zone" {
  description = "Cloudflare zone name (registered domain) that manages DNS"
  type        = string
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID. If empty, looked up by cloudflare_zone."
  type        = string
  default     = ""
}

variable "server_ip" {
  description = "Public IPv4 the A records should point at"
  type        = string
}
