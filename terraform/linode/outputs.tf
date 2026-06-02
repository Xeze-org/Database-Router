############################################
# Normalized outputs — identical across every provider root so the
# deployer can consume them without knowing the cloud.
############################################

output "server_ip" {
  description = "Public IPv4 of the server"
  value       = module.compute.server_ip
}

output "ssh_user" {
  description = "SSH login user for the server's OS image"
  value       = module.compute.ssh_user
}

output "fqdn" {
  description = "Full domain name for the db-router"
  value       = "${var.subdomain}.${var.domain}"
}

output "grpc_endpoint" {
  description = "gRPC connection string"
  value       = "${var.subdomain}.${var.domain}:${var.grpc_port}"
}

output "postgres_password" {
  description = "Auto-generated PostgreSQL password"
  value       = module.secrets.postgres_password
  sensitive   = true
}

output "mongo_password" {
  description = "Auto-generated MongoDB password"
  value       = module.secrets.mongo_password
  sensitive   = true
}

output "redis_password" {
  description = "Auto-generated Redis password"
  value       = module.secrets.redis_password
  sensitive   = true
}

output "credentials_summary" {
  description = "All credentials in one block — run: terraform output credentials_summary"
  value       = module.secrets.credentials_summary
  sensitive   = true
}
