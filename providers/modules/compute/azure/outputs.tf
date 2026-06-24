output "server_ip" {
  description = "Public IPv4 of the Azure VM"
  value       = azurerm_public_ip.this.ip_address
}

output "server_id" {
  description = "Azure VM resource ID"
  value       = azurerm_linux_virtual_machine.this.id
}

output "ssh_user" {
  description = "Admin SSH login user"
  value       = local.ssh_user
}
