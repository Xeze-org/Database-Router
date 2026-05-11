#!/bin/bash
set -e

PROJECT_ID="astralelite"
REGION="asia-south1"
REPO_NAME="dbr-apps"
IMAGE_NAME="gcp-cloudrun-example"
DB_ROUTER_HOST="db.0.xeze.org:443"

IMAGE_URI="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/${IMAGE_NAME}:latest"

# Find directories
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
APP_DIR="$(dirname "$SCRIPT_DIR")/gcp-cloudrun"
CERTS_DIR="$(dirname "$SCRIPT_DIR")/../deployer/state/certs"

echo ""
echo "=========================================================="
echo ">>> Database Router - Automated Cloud Run Deployer"
echo "=========================================================="
echo "Project ID: ${PROJECT_ID}"
echo "Region:     ${REGION}"
echo "Image URI:  ${IMAGE_URI}"
echo "Certs Dir:  ${CERTS_DIR}"
echo "=========================================================="
echo ""

# Validate certs exist
if [ ! -f "$CERTS_DIR/client.crt" ] || [ ! -f "$CERTS_DIR/client.key" ]; then
    echo "[Error] Certificates not found in $CERTS_DIR"
    echo "        Run the deployer first to generate and fetch certs."
    exit 1
fi

# Copy certs to /tmp to avoid WSL mount permission issues on private keys
cp "$CERTS_DIR/client.crt" /tmp/dbr-client.crt 2>/dev/null || sudo cp "$CERTS_DIR/client.crt" /tmp/dbr-client.crt
cp "$CERTS_DIR/client.key" /tmp/dbr-client.key 2>/dev/null || sudo cp "$CERTS_DIR/client.key" /tmp/dbr-client.key
chmod 644 /tmp/dbr-client.crt /tmp/dbr-client.key

CERT_FILE="/tmp/dbr-client.crt"
KEY_FILE="/tmp/dbr-client.key"

echo "[OK] Certificates staged to /tmp"

# ── Step 1: Upload certs to GCP Secret Manager (global) ─────────────

echo ""
echo "[1/5] Uploading certificates to GCP Secret Manager..."

upload_secret() {
    local name=$1
    local file=$2
    
    # Create secret if it doesn't exist
    if ! gcloud secrets describe "$name" --project="$PROJECT_ID" &>/dev/null; then
        echo "  Creating secret: $name"
        gcloud secrets create "$name" --project="$PROJECT_ID" --replication-policy=automatic
    fi
    
    # Add a new version — pipe via cat to avoid WSL file permission issues
    echo "  Uploading new version for: $name"
    cat "$file" | gcloud secrets versions add "$name" --project="$PROJECT_ID" --data-file=-
}

upload_secret "dbr-client-cert" "$CERT_FILE"
upload_secret "dbr-client-key"  "$KEY_FILE"

echo "[OK] Secrets uploaded."

# ── Step 2: Artifact Registry ───────────────────────────────────────

echo ""
echo "[2/5] Checking/Creating Artifact Registry Repository..."
gcloud artifacts repositories create "$REPO_NAME" \
    --repository-format=docker \
    --location="$REGION" \
    --description="Database Router Apps" \
    --project="$PROJECT_ID" 2>/dev/null || true

# ── Step 3: Docker Auth ─────────────────────────────────────────────

echo ""
echo "[3/5] Configuring Docker Authentication..."
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet 2>/dev/null

# ── Step 4: Build & Push ────────────────────────────────────────────

echo ""
echo "[4/5] Building & Pushing Docker Image..."
cd "$APP_DIR"
docker build -t "$IMAGE_URI" .
docker push "$IMAGE_URI"

# ── Step 5: Deploy to Cloud Run ─────────────────────────────────────

echo ""
echo "[5/5] Deploying to Cloud Run..."
gcloud run deploy "$IMAGE_NAME" \
  --image "$IMAGE_URI" \
  --region "$REGION" \
  --project "$PROJECT_ID" \
  --allow-unauthenticated \
  --set-env-vars="DB_ROUTER_HOST=${DB_ROUTER_HOST}" \
  --set-env-vars="DB_ROUTER_CERT=/secrets/cert/client.crt" \
  --set-env-vars="DB_ROUTER_KEY=/secrets/key/client.key" \
  --set-secrets="/secrets/cert/client.crt=dbr-client-cert:latest" \
  --set-secrets="/secrets/key/client.key=dbr-client-key:latest"

echo ""
echo "[Success] Deployment complete!"
echo "Click the Cloud Run URL above to view your application!"
