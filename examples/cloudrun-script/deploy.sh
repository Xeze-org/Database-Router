#!/bin/bash
set -e

# ── Configuration ───────────────────────────────────────────────────
PROJECT_ID="astralelite"
REGION="asia-south1"
REPO_NAME="dbr-apps"
IMAGE_NAME="gcp-cloudrun-example"
DB_ROUTER_HOST="db.0.xeze.org:443"
CERTS_DIR="${1:-$HOME/dbr-certs}"

IMAGE_URI="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/${IMAGE_NAME}:latest"

echo ""
echo "=========================================================="
echo ">>> Database Router - Cloud Run Deployer"
echo "=========================================================="
echo "Project ID: ${PROJECT_ID}"
echo "Region:     ${REGION}"
echo "Image URI:  ${IMAGE_URI}"
echo "Certs Dir:  ${CERTS_DIR}"
echo "=========================================================="
echo ""

# ── Validate certs ──────────────────────────────────────────────────
if [ ! -f "$CERTS_DIR/client.crt" ] || [ ! -f "$CERTS_DIR/client.key" ]; then
    echo "[Error] client.crt or client.key not found in $CERTS_DIR"
    echo ""
    echo "  Copy them from Windows first:"
    echo "    mkdir -p ~/dbr-certs"
    echo "    cp /mnt/e/projects/Database-Router/deployer/state/certs/* ~/dbr-certs/"
    exit 1
fi
echo "[OK] Certificates found."

# ── 1. Upload certs as global secrets ───────────────────────────────
echo ""
echo "[1/5] Uploading certificates to GCP Secret Manager..."

upload_secret() {
    local name=$1
    local file=$2

    if ! gcloud secrets describe "$name" --project="$PROJECT_ID" &>/dev/null; then
        echo "  Creating secret: $name"
        gcloud secrets create "$name" --project="$PROJECT_ID" --replication-policy=automatic
    fi

    echo "  Uploading version for: $name"
    gcloud secrets versions add "$name" --project="$PROJECT_ID" --data-file="$file"
}

upload_secret "dbr-client-cert" "$CERTS_DIR/client.crt"
upload_secret "dbr-client-key"  "$CERTS_DIR/client.key"
echo "[OK] Secrets uploaded."

# ── 2. Artifact Registry ───────────────────────────────────────────
echo ""
echo "[2/5] Ensuring Artifact Registry exists..."
gcloud artifacts repositories create "$REPO_NAME" \
    --repository-format=docker \
    --location="$REGION" \
    --description="Database Router Apps" \
    --project="$PROJECT_ID" 2>/dev/null || true

# ── 3. Docker Auth ─────────────────────────────────────────────────
echo ""
echo "[3/5] Configuring Docker auth..."
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet 2>/dev/null

# ── 4. Build & Push ────────────────────────────────────────────────
echo ""
echo "[4/5] Building and pushing Docker image..."

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
APP_DIR="$(dirname "$SCRIPT_DIR")/gcp-cloudrun"
cd "$APP_DIR"

docker build -t "$IMAGE_URI" .
docker push "$IMAGE_URI"

# ── 5. Deploy ──────────────────────────────────────────────────────
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

if [ $? -ne 0 ]; then
    echo ""
    echo "[Error] Deployment failed!"
    exit 1
fi

echo ""
echo "[Success] Deployment complete!"
echo "Visit the Cloud Run URL above to see your app."
