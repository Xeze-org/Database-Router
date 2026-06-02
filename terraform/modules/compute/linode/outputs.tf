output "server_ip" {
  description = "Public IPv4 of the Linode"
  value       = linode_instance.this.ip_address
}

output "server_id" {
  description = "Linode instance ID"
  value       = linode_instance.this.id
}

output "ssh_user" {
  description = "Default SSH login user for the OS image"
  value       = "root"
}
