############################################
# Firewall — db-router
############################################

resource "digitalocean_firewall" "db_router_fw" {
  name       = "${var.droplet_name}-fw"
  depends_on = [digitalocean_droplet.db_router]

  droplet_ids = [digitalocean_droplet.db_router.id]

  # ── Inbound rules ────────────────────────────────────────────────────

  # SSH — restricted to allowed IPs
  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = var.allowed_ips
  }

  # HTTP — open (required for Let's Encrypt ACME + Caddy redirect)
  inbound_rule {
    protocol         = "tcp"
    port_range       = "80"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # HTTPS — open (Caddy serves Web UI + gRPC through TLS)
  inbound_rule {
    protocol         = "tcp"
    port_range       = "443"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # ── Outbound rules (unrestricted) ──────────────────────────────────

  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}
