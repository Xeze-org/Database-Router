############################################
# Cloudflare DNS — A records for the db-router
#
# Cloud-agnostic: works with any compute provider. The compute module
# hands us a public IP; we point the domain at it.
#
# proxied = false (DNS-only / "grey cloud") is REQUIRED:
#   - Caddy terminates TLS and enforces mTLS at the edge; Cloudflare's
#     proxy would strip client certs and break mTLS.
#   - gRPC runs over h2c behind Caddy; the CF proxy does not pass it cleanly.
#   - Let's Encrypt ACME (HTTP-01) needs a direct path to the origin.
############################################

terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
}

# Resolve the zone id from the zone name unless one was provided explicitly.
data "cloudflare_zone" "this" {
  count = var.cloudflare_zone_id == "" ? 1 : 0
  name  = var.cloudflare_zone
}

locals {
  zone_id = var.cloudflare_zone_id != "" ? var.cloudflare_zone_id : data.cloudflare_zone.this[0].id

  # Record name relative to the zone. If domain == zone, the record is just
  # the subdomain; otherwise it includes the delegated sub-zone segment.
  base_name = var.domain == var.cloudflare_zone ? var.subdomain : "${var.subdomain}.${trimsuffix(var.domain, ".${var.cloudflare_zone}")}"
}

resource "cloudflare_record" "db_router" {
  zone_id = local.zone_id
  name    = local.base_name
  type    = "A"
  content = var.server_ip
  ttl     = 300
  proxied = false
}

resource "cloudflare_record" "grpc" {
  zone_id = local.zone_id
  name    = "grpc.${local.base_name}"
  type    = "A"
  content = var.server_ip
  ttl     = 300
  proxied = false
}
