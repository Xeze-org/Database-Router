############################################
# Linode compute — instance + firewall
############################################

terraform {
  required_providers {
    linode = {
      source  = "linode/linode"
      version = "~> 2.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.6"
    }
  }
}

locals {
  allowed_ipv4 = [for c in var.allowed_ips : c if !strcontains(c, ":")]
  allowed_ipv6 = [for c in var.allowed_ips : c if strcontains(c, ":")]
}

# Linode requires a root password even when an SSH key is provided.
resource "random_password" "root" {
  length           = 28
  special          = true
  override_special = "!@#%^*-_=+"
  min_lower        = 2
  min_upper        = 2
  min_numeric      = 2
}

resource "linode_instance" "this" {
  label           = var.server_name
  region          = var.region
  type            = var.instance_size
  image           = var.image
  authorized_keys = [trimspace(var.ssh_public_key)]
  root_pass       = random_password.root.result
  tags            = var.tags
}

resource "linode_firewall" "this" {
  label           = "${var.server_name}-fw"
  inbound_policy  = "DROP"
  outbound_policy = "ACCEPT"

  inbound {
    label    = "ssh"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "22"
    ipv4     = local.allowed_ipv4
    ipv6     = local.allowed_ipv6
  }
  inbound {
    label    = "http"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "80"
    ipv4     = ["0.0.0.0/0"]
    ipv6     = ["::/0"]
  }
  inbound {
    label    = "https"
    action   = "ACCEPT"
    protocol = "TCP"
    ports    = "443"
    ipv4     = ["0.0.0.0/0"]
    ipv6     = ["::/0"]
  }

  linodes = [linode_instance.this.id]
}
