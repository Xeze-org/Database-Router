############################################
# SSH key — hybrid: existing (mounted) or generated (container)
#
# If ssh_public_key is set → upload it to DO as a new key.
# If empty                 → look up an existing key by ssh_key_name.
############################################

resource "digitalocean_ssh_key" "generated" {
  count      = var.ssh_public_key != "" ? 1 : 0
  name       = "${var.droplet_name}-deploy-key"
  public_key = var.ssh_public_key
}

data "digitalocean_ssh_key" "existing" {
  count = var.ssh_public_key == "" ? 1 : 0
  name  = var.ssh_key_name
}

locals {
  ssh_key_id = (
    var.ssh_public_key != ""
    ? digitalocean_ssh_key.generated[0].id
    : data.digitalocean_ssh_key.existing[0].id
  )
}
