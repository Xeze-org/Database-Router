############################################
# AWS compute — EC2 instance in the default VPC + security group
############################################

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

locals {
  allowed_ipv4 = [for c in var.allowed_ips : c if !strcontains(c, ":")]
  allowed_ipv6 = [for c in var.allowed_ips : c if strcontains(c, ":")]
}

data "aws_vpc" "default" {
  default = true
}

# Latest official Debian 12 (Bookworm) AMI — used when var.image is empty.
data "aws_ami" "debian" {
  most_recent = true
  owners      = ["136693071363"] # Debian

  filter {
    name   = "name"
    values = ["debian-12-amd64-*"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_key_pair" "this" {
  key_name   = "${var.server_name}-key"
  public_key = var.ssh_public_key
}

resource "aws_security_group" "this" {
  name        = "${var.server_name}-sg"
  description = "db-router ingress/egress"
  vpc_id      = data.aws_vpc.default.id
  tags        = { Name = "${var.server_name}-sg" }

  ingress {
    description      = "SSH"
    from_port        = 22
    to_port          = 22
    protocol         = "tcp"
    cidr_blocks      = local.allowed_ipv4
    ipv6_cidr_blocks = local.allowed_ipv6
  }
  ingress {
    description      = "HTTP (ACME)"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
  ingress {
    description      = "HTTPS (Caddy/gRPC)"
    from_port        = 443
    to_port          = 443
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_instance" "this" {
  ami                         = var.image != "" ? var.image : data.aws_ami.debian.id
  instance_type               = var.instance_size
  key_name                    = aws_key_pair.this.key_name
  vpc_security_group_ids      = [aws_security_group.this.id]
  associate_public_ip_address = true

  tags = { Name = var.server_name }
}
