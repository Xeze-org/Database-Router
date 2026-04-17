############################################
# Local aliases for the auto-generated passwords
############################################

locals {
  pg_pass    = random_password.postgres.result
  mongo_pass = random_password.mongo.result
  redis_pass = random_password.redis.result
}

############################################
# Droplet — db-router server
# Ansible handles all provisioning after creation.
############################################

resource "digitalocean_droplet" "db_router" {
  name   = var.droplet_name
  region = var.region
  size   = var.droplet_size
  image  = var.image

  ssh_keys = [local.ssh_key_id]

  tags = ["terraform", "db-router", "grpc"]
}
