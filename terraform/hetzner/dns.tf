############################################
# DNS — A records pointing to the server
#
# Hetzner DNS is managed via the hcloud provider.
# The domain must already exist in your Hetzner DNS console.
############################################

data "hcloud_dns_zone" "base" {
  name = var.domain
}

resource "hcloud_dns_record" "db_router" {
  zone_id = data.hcloud_dns_zone.base.id
  type    = "A"
  name    = var.subdomain
  value   = hcloud_server.db_router.ipv4_address
  ttl     = 300
}

resource "hcloud_dns_record" "grpc" {
  zone_id = data.hcloud_dns_zone.base.id
  type    = "A"
  name    = "grpc.${var.subdomain}"
  value   = hcloud_server.db_router.ipv4_address
  ttl     = 300
}
