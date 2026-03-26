############################################
# DNS — A record pointing to the droplet
############################################

# Use an existing domain already managed in DO
# (if the domain doesn't exist yet, uncomment the resource below)

# resource "digitalocean_domain" "base" {
#   name = var.domain
# }

resource "digitalocean_record" "db_router" {
  domain = var.domain
  type   = "A"
  name   = var.subdomain
  value  = digitalocean_droplet.db_router.ipv4_address
  ttl    = 300
}
