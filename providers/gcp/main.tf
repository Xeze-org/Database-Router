############################################
# gcp root — wires the shared modules.
############################################

module "secrets" {
  source = "../modules/secrets"

  postgres_user = var.postgres_user
  postgres_db   = var.postgres_db
  mongo_user    = var.mongo_user
}

module "compute" {
  source = "../modules/compute/gcp"

  server_name    = var.server_name
  region         = var.region
  instance_size  = var.instance_size
  image          = var.image
  ssh_public_key = var.ssh_public_key
  allowed_ips    = var.allowed_ips
  tags           = var.tags
}

module "dns" {
  source = "../modules/dns-cloudflare"

  domain             = var.domain
  subdomain          = var.subdomain
  cloudflare_zone    = var.cloudflare_zone
  cloudflare_zone_id = var.cloudflare_zone_id
  server_ip          = module.compute.server_ip
}
