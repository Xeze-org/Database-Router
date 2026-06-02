# db-router Deployer

One-command deployment of the full db-router stack on the cloud of your choice — **DigitalOcean, Linode, Hetzner, AWS, GCP, or Azure**. DNS is always managed in **Cloudflare**.

A Docker container bundles Terraform + Ansible and orchestrates:

1. **Terraform** provisions the server, firewall, and Cloudflare DNS for the selected `CLOUD_PROVIDER`
2. **Ansible** SSHes in to install Docker, deploy the app, and configure Caddy (auto-HTTPS)

```
You (laptop)
  │
  ├── docker run -e CLOUD_PROVIDER=hetzner -e HCLOUD_TOKEN=xxx -e CLOUDFLARE_API_TOKEN=cf_xxx
  │       │
  │       ▼
  │   ┌─────────────── Container ──────────────────┐
  │   │  1. Detect/generate SSH key                │
  │   │  2. terraform apply  →  server + CF DNS    │
  │   │  3. ansible-playbook →  Docker + app       │
  │   │  4. Print credentials + SSH config         │
  │   └────────────────────────────────────────────┘
  │
  └── ssh db-router   ← ready to use
```

---

## Quick start

Set `CLOUD_PROVIDER`, that provider's credentials, and `CLOUDFLARE_API_TOKEN`:

```bash
cd deployer

# DigitalOcean (default provider)
CLOUD_PROVIDER=digitalocean DIGITALOCEAN_TOKEN=dop_v1_xxx \
CLOUDFLARE_API_TOKEN=cf_xxx docker compose up

# Hetzner
CLOUD_PROVIDER=hetzner HCLOUD_TOKEN=xxx \
CLOUDFLARE_API_TOKEN=cf_xxx docker compose up

# Done — credentials are printed at the end
```

That's it. No SSH key setup, no config files, no manual steps.

### Credentials per provider

| `CLOUD_PROVIDER` | Required compute env vars |
|---|---|
| `digitalocean` | `DIGITALOCEAN_TOKEN` |
| `linode` | `LINODE_TOKEN` |
| `hetzner` | `HCLOUD_TOKEN` |
| `aws` | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` |
| `gcp` | `GOOGLE_CREDENTIALS`, `GOOGLE_PROJECT` |
| `azure` | `ARM_CLIENT_ID`, `ARM_CLIENT_SECRET`, `ARM_SUBSCRIPTION_ID`, `ARM_TENANT_ID` |

`CLOUDFLARE_API_TOKEN` is **always required** (DNS).

---

## SSH key handling (hybrid)

The container detects the best key source automatically:

| Scenario | What happens |
|---|---|
| You mount `~/.ssh` | Uploads your existing public key to the provider |
| No key mounted, first run | Generates an ed25519 key, uploads the public key |
| No key mounted, re-run | Reuses the key from `state/.ssh/` |

### Use your own key

```bash
docker run --rm -it \
  -e CLOUD_PROVIDER=hetzner \
  -e HCLOUD_TOKEN=xxx \
  -e CLOUDFLARE_API_TOKEN=cf_xxx \
  -v ~/.ssh:/root/.ssh:ro \
  -v $(pwd)/state:/workspace/state \
  db-router-deploy
```

### Let the container generate one

```bash
docker run --rm -it \
  -e CLOUD_PROVIDER=hetzner \
  -e HCLOUD_TOKEN=xxx \
  -e CLOUDFLARE_API_TOKEN=cf_xxx \
  -v $(pwd)/state:/workspace/state \
  db-router-deploy
```

The generated key persists in `state/.ssh/` for future runs.

---

## Configuration

All config is via environment variables. Nothing needs to be hardcoded.

| Variable | Default | Description |
|---|---|---|
| `CLOUD_PROVIDER` | `digitalocean` | `digitalocean` \| `linode` \| `hetzner` \| `aws` \| `gcp` \| `azure` |
| *(provider token)* | **(required)** | See "Credentials per provider" above |
| `CLOUDFLARE_API_TOKEN` | **(required)** | Cloudflare token with `Zone:Read` + `DNS:Edit` |
| `DESTROY` | `false` | Set to `true` to tear down everything |
| `TF_VAR_domain` | `0.xeze.org` | Base domain (managed in Cloudflare) |
| `TF_VAR_subdomain` | `db` | Subdomain for the A record |
| `TF_VAR_cloudflare_zone` | `xeze.org` | Cloudflare zone (registered domain) |
| `TF_VAR_region` | *(per cloud)* | Region / location / zone |
| `TF_VAR_instance_size` | *(per cloud)* | Instance size/type |
| `TF_VAR_enable_mtls` | `true` | Enable mTLS on gRPC |
| `CADDY_EMAIL` | `admin@{domain}` | Email for Let's Encrypt |

---

## State persistence

The `state/` directory (mounted as a Docker volume) stores:

```
state/
├── terraform-<provider>.tfstate   ← infrastructure state (per provider)
├── terraform-<provider>.tfstate.backup
├── certs/                         ← fetched mTLS certs (client.crt/key, ca.crt)
├── ssh-config                     ← ready-to-use SSH config snippet
└── .ssh/                          ← auto-generated key (if no key was mounted)
    ├── id_ed25519
    └── id_ed25519.pub
```

State is namespaced per `CLOUD_PROVIDER` so switching clouds never clobbers it.

**Do not delete `state/`** — without the `.tfstate`, Terraform can't manage or destroy the infrastructure.

---

## Operations

> The examples below use Hetzner; swap `CLOUD_PROVIDER` + the token for any provider.

### Deploy

```bash
CLOUD_PROVIDER=hetzner HCLOUD_TOKEN=xxx CLOUDFLARE_API_TOKEN=cf_xxx docker compose up
```

### Re-deploy (update config / re-run Ansible)

```bash
CLOUD_PROVIDER=hetzner HCLOUD_TOKEN=xxx CLOUDFLARE_API_TOKEN=cf_xxx docker compose up
```

Safe to re-run. Terraform is idempotent; Ansible re-configures only what changed.

### Destroy everything

```bash
CLOUD_PROVIDER=hetzner HCLOUD_TOKEN=xxx CLOUDFLARE_API_TOKEN=cf_xxx DESTROY=true docker compose up
```

Tears down the server, firewall, Cloudflare DNS records, and uploaded SSH key.

### Override domain/region

```bash
docker run --rm -it \
  -e CLOUD_PROVIDER=hetzner \
  -e HCLOUD_TOKEN=xxx \
  -e CLOUDFLARE_API_TOKEN=cf_xxx \
  -e TF_VAR_domain=example.com \
  -e TF_VAR_subdomain=api \
  -e TF_VAR_region=fsn1 \
  -v $(pwd)/state:/workspace/state \
  db-router-deploy
```

---

## What gets deployed

After a successful run, you get:

| Component | Endpoint |
|---|---|
| gRPC (Caddy mTLS) | `db.0.xeze.org:443` |
| SSH | `ssh <ssh_user>@<ip>` or `ssh db-router` |

Internally on the server:
- Docker CE + Compose
- PostgreSQL, MongoDB, Redis (containers)
- db-router gRPC server (container)
- Caddy reverse proxy (system service, auto-HTTPS via Let's Encrypt, mTLS)

---

## Security

- **Provider & Cloudflare tokens**: only in memory as env vars, never written to disk
- **SSH key**: either your own (read-only mount) or generated in persistent state volume
- **DB passwords**: auto-generated by Terraform, passed to Ansible, never hardcoded
- **Firewall**: auto-detects your public IP and restricts SSH/gRPC to it
- **HTTPS**: Caddy auto-obtains Let's Encrypt certs; ports 80/443 are the only public ports

---

## Building manually

If you prefer to build the image yourself:

```bash
# From the repo root
docker build -f deployer/Dockerfile -t db-router-deploy .

# Run
docker run --rm -it \
  -e CLOUD_PROVIDER=hetzner \
  -e HCLOUD_TOKEN=xxx \
  -e CLOUDFLARE_API_TOKEN=cf_xxx \
  -v $(pwd)/deployer/state:/workspace/state \
  db-router-deploy
```

---

## File structure

```
deployer/
├── Dockerfile           ← Alpine + Terraform + Ansible + SSH
├── entrypoint.sh        ← Orchestrator: SSH detect → TF apply → Ansible → summary
├── docker-compose.yml   ← Convenience wrapper
├── .gitignore           ← Ignores state/
└── README.md            ← This file
```
