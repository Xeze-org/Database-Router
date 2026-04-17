############################################
# Local aliases for the auto-generated passwords
############################################

locals {
  pg_pass    = random_password.postgres.result
  mongo_pass = random_password.mongo.result
  redis_pass = random_password.redis.result
}

############################################
# Server — db-router
# Ansible handles all provisioning after creation.
############################################

resource "hcloud_server" "db_router" {
  name        = var.server_name
  server_type = var.server_type
  image       = var.image
  location    = var.location

  ssh_keys = [local.ssh_key_id]

  labels = {
    managed-by = "terraform"
    service    = "db-router"
    protocol   = "grpc"
  }

  firewall_ids = [hcloud_firewall.db_router_fw.id]
}
