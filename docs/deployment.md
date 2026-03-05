# Deployment

---

## Option 1 — Docker (recommended)

The published image contains zero credentials. You always supply `config.json` at runtime.

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config.json:/app/config.json:ro \
  --restart unless-stopped \
  --name db-router \
  ghcr.io/youruser/db-router:latest
```

Check it started:
```bash
docker logs db-router
curl http://localhost:8080/health
```

---

## Option 2 — Docker Compose (router only)

Use [`deploy/docker-compose.yml`](../deploy/docker-compose.yml) if your databases already run elsewhere:

```bash
cp config.example.json config.json
# fill in your credentials
docker compose -f deploy/docker-compose.yml up -d
```

---

## Option 3 — Docker Compose (full stack)

Use [`docker-compose.yml`](../docker-compose.yml) at the repo root to spin up PostgreSQL, MongoDB, Redis, and db-router together in one command:

```bash
cp config.example.json config.json
docker compose up -d
```

The databases will be available on their standard ports (5432, 27017, 6379) and db-router on 8080.

---

## Option 4 — Build from source

```bash
git clone https://github.com/youruser/db-router
cd db-router
cp config.example.json config.json

# build
go build -o db-router ./cmd/

# run
./db-router
```

Windows:
```bat
start.bat
```

---

## Reverse proxy with Caddy

Put Caddy in front to terminate TLS and enforce an API key:

```caddy
db.yourdomain.com {
    @noauth {
        not header X-API-Key your-secret-key
        not path /health
    }
    respond @noauth 401

    reverse_proxy localhost:8080
}
```

After reloading Caddy, all calls must include:
```
X-API-Key: your-secret-key
```

---

## Reverse proxy with nginx

```nginx
server {
    listen 443 ssl;
    server_name db.yourdomain.com;

    # ... your SSL config ...

    location / {
        if ($http_x_api_key != "your-secret-key") {
            return 401;
        }
        proxy_pass http://127.0.0.1:8080;
    }

    location /health {
        proxy_pass http://127.0.0.1:8080;
    }
}
```

---

## Publishing a new Docker image

The GitHub Actions workflow is **manual only**. To publish a new image:

1. Push your changes to `main`
2. Go to **Actions** in your GitHub repo
3. Click **Build & Publish Docker Image**
4. Click **Run workflow**

The image will be pushed to GHCR as:
- `ghcr.io/youruser/db-router:latest`
- `ghcr.io/youruser/db-router:<short-sha>`
