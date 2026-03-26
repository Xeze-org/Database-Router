############################################
# Connection info
############################################

output "droplet_ip" {
  description = "Public IPv4 of the db-router droplet"
  value       = digitalocean_droplet.db_router.ipv4_address
}

output "fqdn" {
  description = "Full domain name for the db-router"
  value       = "${var.subdomain}.${var.domain}"
}

output "grpc_endpoint" {
  description = "gRPC connection string"
  value       = "${var.subdomain}.${var.domain}:${var.grpc_port}"
}



output "ssh_command" {
  description = "SSH into the droplet"
  value       = "ssh root@${digitalocean_droplet.db_router.ipv4_address}"
}

############################################
# Auto-generated credentials
# Run: terraform output -json
############################################

output "postgres_password" {
  description = "Auto-generated PostgreSQL password"
  value       = random_password.postgres.result
  sensitive   = true
}

output "mongo_password" {
  description = "Auto-generated MongoDB password"
  value       = random_password.mongo.result
  sensitive   = true
}

output "redis_password" {
  description = "Auto-generated Redis password"
  value       = random_password.redis.result
  sensitive   = true
}



############################################
# Quick-copy summary (shown at end of apply)
############################################

output "credentials_summary" {
  description = "All credentials in one block — run: terraform output credentials_summary"
  sensitive   = true
  value       = <<-EOT

    ╔══════════════════════════════════════════════════════════╗
    ║  db-router credentials (auto-generated)                 ║
    ╠══════════════════════════════════════════════════════════╣
    ║                                                          ║
    ║  PostgreSQL                                              ║
    ║    user:     ${var.postgres_user}
    ║    password: ${random_password.postgres.result}
    ║    database: ${var.postgres_db}
    ║                                                          ║
    ║  MongoDB                                                 ║
    ║    user:     ${var.mongo_user}
    ║    password: ${random_password.mongo.result}
    ║                                                          ║
    ║  Redis                                                   ║
    ║    password: ${random_password.redis.result}
    ║                                                          ║

    ║                                                          ║
    ╚══════════════════════════════════════════════════════════╝

  EOT
}
