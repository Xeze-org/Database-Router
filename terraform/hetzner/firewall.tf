############################################
# Firewall — db-router
############################################

resource "hcloud_firewall" "db_router_fw" {
  name = "${var.server_name}-fw"

  # ── Inbound rules ────────────────────────────────────────────────────

  # SSH — restricted to allowed IPs
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "22"
    source_ips = var.allowed_ips
  }

  # HTTP — open (required for Let's Encrypt ACME + Caddy redirect)
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "80"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  # HTTPS — open (Caddy serves gRPC through mTLS)
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "443"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  # ── Outbound rules (unrestricted) ──────────────────────────────────

  rule {
    direction       = "out"
    protocol        = "tcp"
    port            = "1-65535"
    destination_ips = ["0.0.0.0/0", "::/0"]
  }

  rule {
    direction       = "out"
    protocol        = "udp"
    port            = "1-65535"
    destination_ips = ["0.0.0.0/0", "::/0"]
  }

  rule {
    direction       = "out"
    protocol        = "icmp"
    destination_ips = ["0.0.0.0/0", "::/0"]
  }
}
