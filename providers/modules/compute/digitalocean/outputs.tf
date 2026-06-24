output "server_ip" {
  description = "Public IPv4 of the droplet"
  value       = digitalocean_droplet.this.ipv4_address
}

output "server_id" {
  description = "Droplet ID"
  value       = digitalocean_droplet.this.id
}

output "ssh_user" {
  description = "Default SSH login user for the OS image"
  value       = "root"
}
