output "server_ip" {
  description = "Public IPv4 of the Hetzner server"
  value       = hcloud_server.this.ipv4_address
}

output "server_id" {
  description = "Hetzner server ID"
  value       = hcloud_server.this.id
}

output "ssh_user" {
  description = "Default SSH login user for the OS image"
  value       = "root"
}
