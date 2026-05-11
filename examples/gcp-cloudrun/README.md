# GCP Cloud Run Example

This is a ready-to-deploy example Node.js container that connects to your Database Router securely using Google Cloud Secret Manager.

Because it uses Cloud Run's native **Secret Volume Mounts**, there is no GCP-specific code inside the app. It simply reads the certificates from the memory-mounted `/secrets/certs` directory.

## Prerequisites

Before deploying, ensure you have initialized your Google Cloud CLI and set your project:
```bash
gcloud auth login
gcloud config set project YOUR_PROJECT_ID
```

*(Note: Replace `YOUR_PROJECT_ID` with your actual Google Cloud project ID!)*

## Step 1: Build & Push the Container

Now that `@xeze/dbr-core` is published to NPM, building is incredibly simple. 

Run these commands from the `examples/gcp-cloudrun` folder:

```bash
cd E:\projects\Database-Router\examples\gcp-cloudrun

# 1. Enable Artifact Registry (if you haven't already)
gcloud services enable artifactregistry.googleapis.com

# 2. Create a repository for the container (run once)
gcloud artifacts repositories create dbr-apps \
    --repository-format=docker \
    --location=asia-south1 \
    --description="Database Router Apps"

# 3. Configure Docker to use gcloud credentials
gcloud auth configure-docker asia-south1-docker.pkg.dev

# 4. Build and tag the image
docker build -t asia-south1-docker.pkg.dev/YOUR_PROJECT_ID/dbr-apps/gcp-cloudrun-example:latest .

# 5. Push the image to Artifact Registry
docker push asia-south1-docker.pkg.dev/YOUR_PROJECT_ID/dbr-apps/gcp-cloudrun-example:latest
```

## Step 2: Deploy to Cloud Run

The magical part happens during deployment. We tell Cloud Run to take your two secrets and mount them as files at `/secrets/certs/client.crt` and `/secrets/certs/client.key`. 

*(Notice how we handle the typo in your secret name `daabase-router-client-cer` by mapping it correctly to `client.crt`!)*

Run this deployment command:

```bash
gcloud run deploy gcp-cloudrun-example \
  --image asia-south1-docker.pkg.dev/YOUR_PROJECT_ID/dbr-apps/gcp-cloudrun-example:latest \
  --region asia-south1 \
  --allow-unauthenticated \
  --set-env-vars="DB_ROUTER_HOST=db.0.xeze.org:443" \
  --set-env-vars="DB_ROUTER_CERT=/secrets/cert/client.crt" \
  --set-env-vars="DB_ROUTER_KEY=/secrets/key/client.key" \
  --set-secrets="/secrets/cert/client.crt=daabase-router-client-cer:latest" \
  --set-secrets="/secrets/key/client.key=database-router-client-key:latest"
```

## Step 3: Test It!

Once the deployment finishes, the terminal will give you a public URL (e.g., `https://gcp-cloudrun-example-xxxxx-el.a.run.app`).

Click that link! You should immediately see a JSON response confirming that PostgreSQL, MongoDB, and Redis all successfully inserted and queried test data.
