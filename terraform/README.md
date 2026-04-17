# Terraform — Cloud Infrastructure

Provisions the db-router server, firewall, SSH keys, and DNS records on your chosen cloud provider.

## Supported Providers

| Provider | Directory | Auth Env Var | Minimum Cost |
|----------|-----------|-------------|-------------|
| **DigitalOcean** | [`digital-ocean/`](./digital-ocean/) | `DIGITALOCEAN_TOKEN` | ~$12/mo |
| **Hetzner Cloud** | [`hetzner/`](./hetzner/) | `HCLOUD_TOKEN` | ~€3.29/mo |

## Usage

Pick your provider and `cd` into it:

```bash
# DigitalOcean
cd terraform/digital-ocean
export DIGITALOCEAN_TOKEN="dop_v1_..."

# OR Hetzner
cd terraform/hetzner
export HCLOUD_TOKEN="..."
```

Then:

```bash
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values

terraform init
terraform plan
terraform apply
```

## What Gets Created

Each provider directory creates the same logical infrastructure:

1. **Server** — 1 VM running Debian with Docker
2. **SSH Key** — uploaded or looked up from your account
3. **Firewall** — SSH (restricted), HTTP/HTTPS (open for Caddy + ACME)
4. **DNS Records** — `db.yourdomain.com` and `grpc.db.yourdomain.com`
5. **Secrets** — Auto-generated passwords for PostgreSQL, MongoDB, Redis

After Terraform finishes, the deployer runs Ansible to install Docker, db-router, and Caddy on the server. The Ansible playbook is **cloud-agnostic** — it works identically on both providers.

## Destroying

```bash
terraform destroy
```
