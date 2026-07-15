
# Deploying Limiter.io to AWS (Terraform + EC2 + Docker + GitHub Actions)

One EC2 host runs the whole stack in Docker: **nginx → landing (Next.js) + api (Go) + consumer + Postgres + Redis + Kafka**.
Pushing to `main` auto-builds and redeploys via `.github/workflows/deploy.yml`.

```
GitHub push → CI → deploy.yml → rsync to EC2 → docker compose build → up -d → /healthz check
```

---

## 1. One-time: provision AWS with Terraform

```bash
# generate a deploy key (used by you AND GitHub Actions)
ssh-keygen -t ed25519 -f ~/.ssh/limiter_deploy -N ""

cd deploy/terraform
terraform init
terraform apply \
  -var "ssh_public_key=$(cat ~/.ssh/limiter_deploy.pub)" \
  -var "aws_region=ap-south-1"          # change region if you like

terraform output   # note public_ip
```

What it creates: EC2 (t3.micro by default, Ubuntu 24.04, 30 GB gp3), Elastic IP, security group
(22/80/443), and a first-boot script that sets up 4GB swap space (essential to run Kafka + DBs on t3.micro),
and installs Docker + nginx (reverse proxy: `/` → landing:3000, `/api|/swagger|/healthz|/metrics` → api:8080,
with WebSocket upgrade).

---

## 2. Your two .env files — what happens to them

### Root `.env` (backend) — DO NOT ship the dev file as-is
The repo's `.env` is your **dev** file (Neon DB, Upstash Redis, localhost Kafka). In production
the compose stack runs **its own Postgres/Redis/Kafka containers** and force-overrides
`DB_HOST/REDIS_*/KAFKA_BROKERS`, so managed Neon/Upstash are **not needed** on EC2
(one less monthly bill; keep them for local dev).

Create the **prod** version and paste it into the `PROD_ENV_FILE` GitHub secret:

```env
ENV=production
PORT=8080

# In-cluster DB (compose overrides host/port; user+password+name are used to INIT postgres)
DB_USER=postgres
DB_PASSWORD=<generate: openssl rand -hex 24>
DB_NAME=ratelimiter

# JWT — MUST change, the dev value is a known default
JWT_SECRET=<generate: openssl rand -hex 32>
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

ADMIN_EMAIL=admin@ratelimiter.io
ADMIN_PASSWORD=<strong password>

KAFKA_TOPIC=api_logs
KAFKA_GROUP_ID=analytics_consumers

LEMON_SQUEEZY_WEBHOOK_SECRET=my_lemon_webhook_secret_123!
LEMON_SQUEEZY_PRO_VARIANT_ID=1899978

RESEND_API_KEY=<your re_... key>
RESEND_FROM=Limiter.io <onboarding@resend.dev>

# Public URL of the site (reset-password links + CORS)
APP_BASE_URL=http://<EIP-or-domain>
CORS_ALLOWED_ORIGINS=http://<EIP-or-domain>
```

⚠️ Never commit this. The workflow writes it to `/opt/limiter/.env.prod` on the host.

### `landing/.env.local` (frontend) — becomes GitHub secrets
`NEXT_PUBLIC_*` values are **baked in at build time** (they can't be changed at runtime),
so they're passed as Docker build args by the workflow from these secrets:

| GitHub Secret | Value |
|---|---|
| `NEXT_PUBLIC_API_URL` | `http://<EIP-or-domain>/api/v1` ← NOT localhost! |
| `NEXT_PUBLIC_LEMONSQUEEZY_CHECKOUT_URL` | your buy link |
| `NEXT_PUBLIC_SALES_EMAIL` | optional |

Keep `landing/.env.local` for local dev only — it's gitignored.

---

## 3. GitHub repo secrets (Settings → Secrets → Actions)

| Secret | Value |
|---|---|
| `EC2_HOST` | Elastic IP from `terraform output public_ip` |
| `EC2_SSH_KEY` | contents of `~/.ssh/limiter_deploy` (the PRIVATE key) |
| `PROD_ENV_FILE` | the full prod env block from step 2 |
| `NEXT_PUBLIC_API_URL` | `http://<EIP>/api/v1` |
| `NEXT_PUBLIC_LEMONSQUEEZY_CHECKOUT_URL` | your Lemon Squeezy link |
| `NEXT_PUBLIC_SALES_EMAIL` | optional |

---

## 4. Deploy

```bash
git push origin main     # that's it
```

The workflow: rsyncs the repo → writes `.env.prod` → `docker compose -f docker-compose.prod.yml build` →
`up -d` → waits for `http://<EIP>/healthz` = 200. Also runnable manually (workflow_dispatch).

Verify:
- Landing: `http://<EIP>/`
- API: `http://<EIP>/api/v1/...`, Swagger: `http://<EIP>/swagger/index.html`
- Update the **Lemon Squeezy webhook URL** to `http://<EIP>/api/v1/billing/webhook`

---

## 5. Later upgrades (when you're ready)
- **Domain + HTTPS**: point DNS at the EIP, then `sudo apt install certbot python3-certbot-nginx && sudo certbot --nginx`.
  Update `APP_BASE_URL`, `CORS_ALLOWED_ORIGINS`, `NEXT_PUBLIC_API_URL` secrets to `https://yourdomain` and redeploy.
- **Managed data**: swap container Postgres → RDS/Neon and Redis → ElastiCache/Upstash by putting their
  hosts in `PROD_ENV_FILE` and deleting the overrides in `docker-compose.prod.yml`.
- **Scale out**: the `deploy/kubernetes/` manifests are ready when one box isn't enough.

## Troubleshooting
```bash
ssh -i ~/.ssh/limiter_deploy ubuntu@<EIP>
cd /opt/limiter/app/deploy/docker
docker compose -f docker-compose.prod.yml ps         # status
docker compose -f docker-compose.prod.yml logs api   # api logs
sudo systemctl status nginx                          # proxy
```
