############################################
# GCP compute — Compute Engine instance on the default network
#
# Note: var.region carries the GCP *zone* (e.g. us-central1-a) since this
# is a single zonal instance. var.tags maps to GCP network tags.
############################################

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

locals {
  # GCP default network is IPv4-only — drop any IPv6 CIDRs.
  allowed_ipv4 = [for c in var.allowed_ips : c if !strcontains(c, ":")]
  ssh_user     = "deploy"
}

resource "google_compute_firewall" "ssh" {
  name    = "${var.server_name}-ssh"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = local.allowed_ipv4
  target_tags   = ["db-router"]
}

resource "google_compute_firewall" "web" {
  name    = "${var.server_name}-web"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["db-router"]
}

resource "google_compute_instance" "this" {
  name         = var.server_name
  machine_type = var.instance_size
  zone         = var.region
  tags         = ["db-router"]

  boot_disk {
    initialize_params {
      image = var.image
    }
  }

  network_interface {
    network = "default"
    access_config {} # ephemeral public IP
  }

  metadata = {
    ssh-keys = "${local.ssh_user}:${trimspace(var.ssh_public_key)}"
  }
}
