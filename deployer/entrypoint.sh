#!/usr/bin/env bash
set -euo pipefail

# ──────────────────────────────────────────────────────────────────────────────
# db-router deployer — fully automated Terraform + Ansible orchestrator
# ──────────────────────────────────────────────────────────────────────────────

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

STATE_DIR="/workspace/state"
TF_DIR="/workspace/terraform"
ANSIBLE_DIR="/workspace/ansible"

banner() {
  echo -e "${CYAN}"
  echo "  ╔══════════════════════════════════════════════════╗"
  echo "  ║          db-router  deployer                     ║"
  echo "  ║   Terraform → Ansible → Done                    ║"
  echo "  ╚══════════════════════════════════════════════════╝"
  echo -e "${NC}"
}

log()  { echo -e "${GREEN}[deploy]${NC} $*"; }
warn() { echo -e "${YELLOW}[warn]${NC}  $*"; }
die()  { echo -e "${RED}[error]${NC} $*" >&2; exit 1; }

# ── Preflight checks ─────────────────────────────────────────────────────────

banner

[ -z "${DIGITALOCEAN_TOKEN:-}" ] && die "DIGITALOCEAN_TOKEN is not set. Pass it with -e DIGITALOCEAN_TOKEN=..."

mkdir -p "$STATE_DIR"

# ── 1. Hybrid SSH key detection ──────────────────────────────────────────────

log "Detecting SSH keys..."

PRIVATE_KEY=""

if [ -f /root/.ssh/id_ed25519 ]; then
  PRIVATE_KEY="/root/.ssh/id_ed25519"
  log "Found mounted key: $PRIVATE_KEY"
elif [ -f /root/.ssh/id_rsa ]; then
  PRIVATE_KEY="/root/.ssh/id_rsa"
  log "Found mounted key: $PRIVATE_KEY"
elif [ -f "$STATE_DIR/.ssh/id_ed25519" ]; then
  PRIVATE_KEY="$STATE_DIR/.ssh/id_ed25519"
  log "Reusing previously generated key: $PRIVATE_KEY"
else
  log "No SSH key found — generating a new ed25519 key pair..."
  mkdir -p "$STATE_DIR/.ssh"
  ssh-keygen -t ed25519 -f "$STATE_DIR/.ssh/id_ed25519" -N "" -q
  PRIVATE_KEY="$STATE_DIR/.ssh/id_ed25519"
  log "Generated key: $PRIVATE_KEY"
fi

chmod 600 "$PRIVATE_KEY"

if [ -f "${PRIVATE_KEY}.pub" ]; then
  PUB_KEY_CONTENT="$(cat "${PRIVATE_KEY}.pub")"
else
  PUB_KEY_CONTENT=""
fi

if [ -z "${TF_VAR_ssh_key_name:-}" ] && [ -z "$PUB_KEY_CONTENT" ]; then
  die "Mounted SSH key has no .pub file and TF_VAR_ssh_key_name is not set. Cannot proceed."
fi

if [ -n "$PUB_KEY_CONTENT" ] && [ -z "${TF_VAR_ssh_key_name:-}" ]; then
  export TF_VAR_ssh_public_key="$PUB_KEY_CONTENT"
  log "SSH mode: upload public key to DigitalOcean"
else
  export TF_VAR_ssh_public_key=""
  log "SSH mode: lookup existing key '${TF_VAR_ssh_key_name}' in DigitalOcean"
fi

# ── 2. Auto-detect public IP for firewall ────────────────────────────────────

if [ -z "${TF_VAR_allowed_ips:-}" ]; then
  log "Detecting your public IP for firewall rules..."
  MY_IP=$(curl -s --max-time 10 ifconfig.me || curl -s --max-time 10 api.ipify.org || echo "")
  if [ -n "$MY_IP" ]; then
    export TF_VAR_allowed_ips="[\"${MY_IP}/32\"]"
    log "Firewall will allow: ${MY_IP}/32"
  else
    warn "Could not detect public IP — firewall will use default (open). Set TF_VAR_allowed_ips to restrict."
  fi
fi

# ── 3. Terraform ─────────────────────────────────────────────────────────────

cd "$TF_DIR"

TF_STATE="$STATE_DIR/terraform.tfstate"

log "Running terraform init..."
terraform init -input=false

# Destroy mode
if [ "${DESTROY:-false}" = "true" ]; then
  warn "DESTROY mode — tearing down all infrastructure..."
  terraform destroy -auto-approve -input=false -state="$TF_STATE"
  log "Infrastructure destroyed."
  exit 0
fi

log "Running terraform apply..."
terraform apply -auto-approve -input=false -state="$TF_STATE"

# ── 4. Extract Terraform outputs ─────────────────────────────────────────────

log "Extracting Terraform outputs..."

TF_OUT=$(terraform output -json -state="$TF_STATE")

DROPLET_IP=$(echo "$TF_OUT"  | jq -r '.droplet_ip.value')
FQDN=$(echo "$TF_OUT"       | jq -r '.fqdn.value')
PG_PASS=$(echo "$TF_OUT"    | jq -r '.postgres_password.value')
MONGO_PASS=$(echo "$TF_OUT"  | jq -r '.mongo_password.value')
REDIS_PASS=$(echo "$TF_OUT"  | jq -r '.redis_password.value')

[ "$DROPLET_IP" = "null" ] && die "Failed to get droplet IP from Terraform outputs"

log "Droplet IP: $DROPLET_IP"
log "FQDN:       $FQDN"

# ── 5. Wait for SSH to be ready ──────────────────────────────────────────────

log "Waiting for SSH on $DROPLET_IP..."

SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=5 -o LogLevel=ERROR"

for i in $(seq 1 30); do
  if ssh $SSH_OPTS -i "$PRIVATE_KEY" root@"$DROPLET_IP" "echo ok" >/dev/null 2>&1; then
    log "SSH is ready."
    break
  fi
  if [ "$i" -eq 30 ]; then
    die "SSH timed out after 150s. Check that the droplet is running and the SSH key is correct."
  fi
  sleep 5
done

# ── 6. Generate dynamic Ansible inventory ────────────────────────────────────

log "Generating Ansible inventory and variables..."

cat > "$ANSIBLE_DIR/inventory.ini" <<EOF
[dbrouter]
${DROPLET_IP} ansible_user=root ansible_ssh_private_key_file=${PRIVATE_KEY} ansible_ssh_common_args='${SSH_OPTS}'

[dbrouter:vars]
ansible_python_interpreter=/usr/bin/python3
EOF

# Read optional vars from environment with sensible defaults
A_DOMAIN="${TF_VAR_domain:-0.xeze.org}"
A_SUBDOMAIN="${TF_VAR_subdomain:-db}"
A_PG_USER="${TF_VAR_postgres_user:-admin}"
A_PG_DB="${TF_VAR_postgres_db:-unified_db}"
A_MONGO_USER="${TF_VAR_mongo_user:-admin}"
A_MTLS="${TF_VAR_enable_mtls:-true}"
A_CADDY_EMAIL="${CADDY_EMAIL:-admin@${A_DOMAIN}}"

cat > "$ANSIBLE_DIR/group_vars/dbrouter.yml" <<EOF
---
domain: "${A_SUBDOMAIN}.${A_DOMAIN}"
grpc_port: ${TF_VAR_grpc_port:-50051}


postgres_user: "${A_PG_USER}"
postgres_password: "${PG_PASS}"
postgres_db: "${A_PG_DB}"

mongo_user: "${A_MONGO_USER}"
mongo_password: "${MONGO_PASS}"

redis_password: "${REDIS_PASS}"



enable_mtls: ${A_MTLS}
caddy_email: "${A_CADDY_EMAIL}"
EOF

# ── 7. Run Ansible ───────────────────────────────────────────────────────────

log "Running Ansible playbook..."

cd "$ANSIBLE_DIR"
ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook \
  -i inventory.ini \
  playbook.yml

# ── 8. Summary ───────────────────────────────────────────────────────────────

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║${NC}  ${BOLD}db-router deployed successfully${NC}                             ${CYAN}║${NC}"
echo -e "${CYAN}╠══════════════════════════════════════════════════════════════╣${NC}"
echo -e "${CYAN}║${NC}                                                              ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}  ${BOLD}Endpoints${NC}                                                   ${CYAN}║${NC}"

echo -e "${CYAN}║${NC}    gRPC:     grpc.${FQDN}:443                               ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}    SSH:      ssh root@${DROPLET_IP}                          ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}                                                              ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}  ${BOLD}Credentials${NC}                                                 ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}    PostgreSQL: ${A_PG_USER} / ${PG_PASS}                     ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}    MongoDB:    ${A_MONGO_USER} / ${MONGO_PASS}               ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}    Redis:      ${REDIS_PASS}                                 ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}                                                              ${CYAN}║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# ── 9. SSH config snippet ────────────────────────────────────────────────────

echo -e "${YELLOW}Add this to ~/.ssh/config for easy access:${NC}"
echo ""
echo "  Host db-router"
echo "    HostName ${DROPLET_IP}"
echo "    User root"
echo "    IdentityFile ${PRIVATE_KEY}"
echo ""
echo -e "${GREEN}Then connect with:${NC} ssh db-router"
echo ""

# Persist the SSH config snippet to state volume
cat > "$STATE_DIR/ssh-config" <<EOF
Host db-router
  HostName ${DROPLET_IP}
  User root
  IdentityFile ${PRIVATE_KEY}
EOF

log "SSH config saved to state/ssh-config"
log "Done."
