output "server_ip" {
  description = "Public IPv4 of the EC2 instance"
  value       = aws_instance.this.public_ip
}

output "server_id" {
  description = "EC2 instance ID"
  value       = aws_instance.this.id
}

output "ssh_user" {
  description = "Default SSH login user for the Debian AMI"
  value       = "admin"
}
