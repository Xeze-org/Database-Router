output "server_ip" {
  description = "Public IPv4 of the Compute Engine instance"
  value       = google_compute_instance.this.network_interface[0].access_config[0].nat_ip
}

output "server_id" {
  description = "Compute Engine instance ID"
  value       = google_compute_instance.this.instance_id
}

output "ssh_user" {
  description = "SSH login user provisioned via instance metadata"
  value       = local.ssh_user
}
