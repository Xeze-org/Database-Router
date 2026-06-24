#!/usr/bin/env bash
set -euo pipefail

# ──────────────────────────────────────────────────────────────────────────────
# db-router deployer — multi-cloud Terraform + Ansible orchestrator
#
#   CLOUD_PROVIDER selects the compute backend (default: digitalocean).
#   DNS is always managed in Cloudflare (CLOUDFLARE_API_TOKEN required).
# ──────────────────────────────────────────────────────────────────────────────

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

STATE_DIR="/workspace/state"
ANSIBLE_DIR="/workspace/ansible"

CLOUD_PROVIDER="${CLOUD_PROVIDER:-digitalocean}"
TF_DIR="/workspace/providers/${CLOUD_PROVIDER}"

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

# ── Preflight: provider selection + credentials ───────────────────────────────

banner

[ -d "$TF_DIR" ] || die "Unknown CLOUD_PROVIDER '${CLOUD_PROVIDER}'. Valid: digitalocean, linode, hetzner, aws, gcp, azure."

log "Cloud provider: ${BOLD}${CLOUD_PROVIDER}${NC}"

# Required credential env vars per provider (DNS is always Cloudflare).
case "$CLOUD_PROVIDER" in
  digitalocean) REQUIRED_VARS="DIGITALOCEAN_TOKEN" ;;
  linode)       REQUIRED_VARS="LINODE_TOKEN" ;;
  hetzner)      REQUIRED_VARS="HCLOUD_TOKEN" ;;
  aws)          REQUIRED_VARS="AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY" ;;
  gcp)          REQUIRED_VARS="GOOGLE_CREDENTIALS GOOGLE_PROJECT" ;;
  azure)        REQUIRED_VARS="ARM_CLIENT_ID ARM_CLIENT_SECRET ARM_SUBSCRIPTION_ID ARM_TENANT_ID" ;;
  *)            die "Unsupported CLOUD_PROVIDER '${CLOUD_PROVIDER}'." ;;
esac

for v in $REQUIRED_VARS CLOUDFLARE_API_TOKEN; do
  eval "val=\${$v:-}"
  [ -z "$val" ] && die "$v is not set — required for CLOUD_PROVIDER=${CLOUD_PROVIDER}. Pass it with -e $v=..."
done

mkdir -p "$STATE_DIR"

# ── 1. SSH key detection / generation ─────────────────────────────────────────

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

if [ ! -f "${PRIVATE_KEY}.pub" ]; then
  die "SSH key ${PRIVATE_KEY} has no matching .pub file. Cannot upload a public key to the cloud provider."
fi

# Every provider module authorizes this public key on the server.
export TF_VAR_ssh_public_key="$(cat "${PRIVATE_KEY}.pub")"
log "SSH mode: upload public key to ${CLOUD_PROVIDER}"

# ── 2. Auto-detect public IP for firewall ─────────────────────────────────────

if [ -z "${TF_VAR_allowed_ips:-}" ]; then
  log "Detecting your public IP for firewall rules..."
  MY_IP=$(curl -s --max-time 10 ifconfig.me || curl -s --max-time 10 api.ipify.org || echo "")
  if [ -n "$MY_IP" ]; then
    export TF_VAR_allowed_ips="[\"${MY_IP}/32\"]"
    log "Firewall will allow SSH from: ${MY_IP}/32"
  else
    warn "Could not detect public IP — firewall will use default (open). Set TF_VAR_allowed_ips to restrict."
  fi
fi

# ── 3. Terraform ──────────────────────────────────────────────────────────────

cd "$TF_DIR"

# State is namespaced per provider so switching clouds never clobbers it.
TF_STATE="$STATE_DIR/terraform-${CLOUD_PROVIDER}.tfstate"

log "Running terraform init (${CLOUD_PROVIDER})..."
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

# ── 4. Extract Terraform outputs (normalized across all providers) ────────────

log "Extracting Terraform outputs..."

TF_OUT=$(terraform output -json -state="$TF_STATE")

SERVER_IP=$(echo "$TF_OUT"  | jq -r '.server_ip.value')
SSH_USER=$(echo "$TF_OUT"   | jq -r '.ssh_user.value')
FQDN=$(echo "$TF_OUT"       | jq -r '.fqdn.value')
PG_PASS=$(echo "$TF_OUT"    | jq -r '.postgres_password.value')
MONGO_PASS=$(echo "$TF_OUT" | jq -r '.mongo_password.value')
REDIS_PASS=$(echo "$TF_OUT" | jq -r '.redis_password.value')

if [ "$SERVER_IP" = "null" ] || [ -z "$SERVER_IP" ]; then
  die "Failed to get server IP from Terraform outputs"
fi
if [ "$SSH_USER" = "null" ] || [ -z "$SSH_USER" ]; then
  SSH_USER="root"
fi

log "Server IP: $SERVER_IP"
log "SSH user:  $SSH_USER"
log "FQDN:      $FQDN"

# sudo prefix for privileged remote commands (non-root login users)
if [ "$SSH_USER" = "root" ]; then SUDO=""; else SUDO="sudo"; fi

# ── 5. Wait for SSH to be ready ───────────────────────────────────────────────

log "Waiting for SSH on $SERVER_IP..."

SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=5 -o LogLevel=ERROR"

for i in $(seq 1 30); do
  if ssh $SSH_OPTS -i "$PRIVATE_KEY" "${SSH_USER}@${SERVER_IP}" "echo ok" >/dev/null 2>&1; then
    log "SSH is ready."
    break
  fi
  if [ "$i" -eq 30 ]; then
    die "SSH timed out after 150s. Check that the server is running and the SSH key is correct."
  fi
  sleep 5
done

# ── 6. Generate dynamic Ansible inventory ─────────────────────────────────────

log "Generating Ansible inventory and variables..."

cat > "$ANSIBLE_DIR/inventory.ini" <<EOF
[dbrouter]
${SERVER_IP} ansible_user=${SSH_USER} ansible_ssh_private_key_file=${PRIVATE_KEY} ansible_ssh_common_args='${SSH_OPTS}'

[dbrouter:vars]
ansible_python_interpreter=/usr/bin/python3
ansible_become=true
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

# ── 7. Run Ansible ────────────────────────────────────────────────────────────

log "Running Ansible playbook..."

cd "$ANSIBLE_DIR"
ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook \
  -i inventory.ini \
  playbook.yml

# ── 8. Fetch certificates ─────────────────────────────────────────────────────

if [ "$A_MTLS" = "true" ]; then
  log "Fetching mTLS certificates for local testing..."
  mkdir -p "$STATE_DIR/certs"
  # certs are root-owned on the server; bundle them with sudo, then pull.
  if ssh $SSH_OPTS -i "$PRIVATE_KEY" "${SSH_USER}@${SERVER_IP}" \
       "$SUDO tar -czf /tmp/dbr-certs.tgz -C /opt/db-router/certs . && $SUDO chmod 644 /tmp/dbr-certs.tgz" >/dev/null 2>&1; then
    scp $SSH_OPTS -i "$PRIVATE_KEY" "${SSH_USER}@${SERVER_IP}":/tmp/dbr-certs.tgz "$STATE_DIR/dbr-certs.tgz" \
      && tar -xzf "$STATE_DIR/dbr-certs.tgz" -C "$STATE_DIR/certs" \
      && rm -f "$STATE_DIR/dbr-certs.tgz" \
      && log "Certificates saved to $STATE_DIR/certs" \
      || warn "Failed to extract certificates"
  else
    warn "Failed to bundle certificates on the server"
  fi
fi

# ── 9. Summary ────────────────────────────────────────────────────────────────

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║${NC}  ${BOLD}db-router deployed successfully${NC}  (${CLOUD_PROVIDER})            ${CYAN}║${NC}"
echo -e "${CYAN}╠══════════════════════════════════════════════════════════════╣${NC}"
echo -e "${CYAN}║${NC}                                                              ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}  ${BOLD}Endpoints${NC}                                                   ${CYAN}║${NC}"

echo -e "${CYAN}║${NC}    gRPC:     ${FQDN}:443                                    ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}    SSH:      ssh ${SSH_USER}@${SERVER_IP}                    ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}                                                              ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}  ${BOLD}Credentials${NC}                                                 ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}    PostgreSQL: ${A_PG_USER} / ${PG_PASS}                     ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}    MongoDB:    ${A_MONGO_USER} / ${MONGO_PASS}               ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}    Redis:      ${REDIS_PASS}                                 ${CYAN}║${NC}"
echo -e "${CYAN}║${NC}                                                              ${CYAN}║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# ── 10. SSH config snippet ────────────────────────────────────────────────────

echo -e "${YELLOW}Add this to ~/.ssh/config for easy access:${NC}"
echo ""
echo "  Host db-router"
echo "    HostName ${SERVER_IP}"
echo "    User ${SSH_USER}"
echo "    IdentityFile ${PRIVATE_KEY}"
echo ""
echo -e "${GREEN}Then connect with:${NC} ssh db-router"
echo ""

# Persist the SSH config snippet to state volume
cat > "$STATE_DIR/ssh-config" <<EOF
Host db-router
  HostName ${SERVER_IP}
  User ${SSH_USER}
  IdentityFile ${PRIVATE_KEY}
EOF

log "SSH config saved to state/ssh-config"
log "Done."
