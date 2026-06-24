# db-router — Terraform (Multi-Cloud)

Deploy the full db-router stack (PostgreSQL + MongoDB + Redis + gRPC router) on a single VM, on the cloud of your choice. DNS is always managed in **Cloudflare**.

> **Prefer the automated deployer?** See [deployer/](../deployer/) — one `docker run` handles Terraform + Ansible end-to-end with zero manual steps.

## Supported providers

| Provider | Root dir | Compute credentials (env) | Defaults (region / size / image) |
|---|---|---|---|
| DigitalOcean | [`digitalocean/`](digitalocean/) | `DIGITALOCEAN_TOKEN` | `blr1` / `s-1vcpu-2gb` / `debian-13-x64` |
| Linode | [`linode/`](linode/) | `LINODE_TOKEN` | `ap-south` / `g6-nanode-1` / `linode/debian12` |
| Hetzner | [`hetzner/`](hetzner/) | `HCLOUD_TOKEN` | `nbg1` / `cx22` / `debian-12` |
| AWS | [`aws/`](aws/) | `AWS_ACCESS_KEY_ID` + `AWS_SECRET_ACCESS_KEY` | `us-east-1` / `t3.small` / latest Debian 12 AMI |
| GCP | [`gcp/`](gcp/) | `GOOGLE_CREDENTIALS` + `GOOGLE_PROJECT` | `us-central1-a` / `e2-small` / `debian-cloud/debian-12` |
| Azure | [`azure/`](azure/) | `ARM_CLIENT_ID` + `ARM_CLIENT_SECRET` + `ARM_SUBSCRIPTION_ID` + `ARM_TENANT_ID` | `eastus` / `Standard_B1ms` / Debian 12 |

**DNS (all providers):** `CLOUDFLARE_API_TOKEN`.

---

## Architecture

Each provider has a thin **root directory** that wires three shared modules:

```
providers/
├── modules/
│   ├── secrets/          # auto-generated DB passwords + credentials summary
│   ├── dns-cloudflare/   # Cloudflare A records (DNS-only / proxied=false)
│   └── compute/<cloud>/  # the VM + firewall/security-group + SSH key
└── <cloud>/              # root: provider blocks + module wiring + outputs
```

Every `compute/<cloud>` module exposes the **same contract** — inputs
(`server_name`, `region`, `instance_size`, `image`, `ssh_public_key`,
`allowed_ips`, `tags`) and outputs (`server_ip`, `server_id`, `ssh_user`) — so
the roots are near-identical and the deployer is cloud-agnostic.

All roots expose the same **normalized outputs**: `server_ip`, `ssh_user`,
`fqdn`, `grpc_endpoint`, `postgres_password`, `mongo_password`,
`redis_password`, `credentials_summary`.

---

## Prerequisites

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.5
- An account + API credentials for your chosen cloud (see table above)
- A domain whose DNS is managed in **Cloudflare** + a `CLOUDFLARE_API_TOKEN`
  with `Zone:Read` and `DNS:Edit` on that zone
- An SSH key (or use the deployer container, which auto-generates one)

---

## Quick start

Pick your provider's directory and set the matching credentials. Example for **Hetzner**:

```bash
export HCLOUD_TOKEN="xxxxxxxx"
export CLOUDFLARE_API_TOKEN="cf_xxxxxxxx"

cd providers/hetzner
cp terraform.tfvars.example terraform.tfvars   # edit allowed_ips, domain, ssh_public_key
terraform init
terraform plan
terraform apply
```

DigitalOcean instead? `cd providers/digitalocean` and `export DIGITALOCEAN_TOKEN=...`. The flow is identical for every provider.

Then run [Ansible](../ansible/) to configure the server, or use the [deployer container](../deployer/) which does both automatically.

### Verify

```bash
ssh $(terraform output -raw ssh_user)@$(terraform output -raw server_ip)

grpcurl \
  -cacert certs/ca.crt \
  -cert   certs/client.crt \
  -key    certs/client.key \
  $(terraform output -raw fqdn):443 dbrouter.HealthService/Check
```

---

## Common variables

These exist in every root (provider-specific compute defaults differ — see the table):

| Variable | Default | Description |
|---|---|---|
| `server_name` | `db-router` | Instance name |
| `region` | *(per cloud)* | Region / location / zone |
| `instance_size` | *(per cloud)* | Instance size/type |
| `image` | *(per cloud)* | OS image (Debian 12/13) |
| `ssh_public_key` | `""` | Public key uploaded to the server (required for manual use) |
| `domain` | `0.xeze.org` | Base domain |
| `subdomain` | `db` | Subdomain (→ `db.0.xeze.org`) |
| `cloudflare_zone` | `xeze.org` | Cloudflare zone (registered domain) managing DNS |
| `cloudflare_zone_id` | `""` | Zone ID; looked up by `cloudflare_zone` when empty |
| `postgres_user` | `admin` | PG username |
| `postgres_db` | `unified_db` | Default PG database |
| `mongo_user` | `admin` | Mongo username |
| `allowed_ips` | `["0.0.0.0/0","::/0"]` | CIDRs allowed to reach SSH — **set to your IP!** |
| `grpc_port` | `50051` | gRPC port (internal; fronted by Caddy on 443) |

> **Passwords are auto-generated.** View them with `terraform output credentials_summary`.

---

## DNS notes (Cloudflare)

Records are created **DNS-only (`proxied = false`, grey cloud)** on purpose:
Caddy terminates TLS and enforces mTLS at the edge, gRPC runs over h2c, and
Let's Encrypt needs a direct path to the origin — Cloudflare's proxy would
break all three. The module creates `A` records for `<subdomain>.<domain>` and
`grpc.<subdomain>.<domain>`.

---

## Security hardening (production)

1. **Set `allowed_ips`** to your IP/office CIDR — e.g. `["203.0.113.5/32"]`.
2. **Enable mTLS** (the deployer sets `enable_mtls = true`) — only clients with a signed cert can connect.
3. **Remote state**: add an `s3`/`gcs`/`azurerm` backend for team usage.

---

## Destroy

```bash
cd providers/<provider>
terraform destroy
```
