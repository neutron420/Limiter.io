# Step-by-Step AWS Deployment Guide (t3.micro Free Tier)

Follow these exact steps tomorrow to deploy Limiter.io to AWS using your free tier credits.

---

## Step 1: Install & Configure AWS CLI
If you haven't installed it, download and install the AWS CLI first.

1. **Create an IAM User on AWS**:
   - Open the **AWS Console**.
   - Search for **IAM** -> click **Users** -> **Create user**.
   - Set user name: `limiter-admin`.
   - Click **Next** -> Choose **Attach policies directly**.
   - Search and check **`AdministratorAccess`** (this allows Terraform to create VPC, Security Groups, and EC2 instances).
   - Click **Next** -> **Create user**.

2. **Generate Access Keys**:
   - Click on your newly created user (`limiter-admin`).
   - Go to the **Security credentials** tab.
   - Scroll down to **Access keys** -> click **Create access key**.
   - Select **Command Line Interface (CLI)**.
   - Acknowledge the recommendation, click **Next**, and click **Create access key**.
   - **Copy the Access Key ID and Secret Access Key** (Save them somewhere safe).

3. **Configure AWS CLI locally**:
   Open PowerShell/Terminal on your computer and run:
   ```bash
   aws configure
   ```
   Provide the credentials when prompted:
   - **AWS Access Key ID**: `[Your Access Key ID]`
   - **AWS Secret Access Key**: `[Your Secret Access Key]`
   - **Default region name**: `ap-south-1` (Mumbai, or use `us-east-1`, `us-west-2` etc.)
   - **Default output format**: `json`

---

## Step 2: Generate SSH Deployment Keys (Windows PowerShell)
We need an SSH key pair to allow you and GitHub Actions to log in securely to the EC2 instance.

1. Run this command in Windows PowerShell to generate the key:
   ```powershell
   ssh-keygen -t ed25519 -f "$HOME/.ssh/limiter_deploy" -N ""
   ```
2. This generates two files in `C:\Users\R.K Singh\.ssh\`:
   - `limiter_deploy` (Private Key - keep this secret, we will give this to GitHub Secrets)
   - `limiter_deploy.pub` (Public Key - we will give this to AWS/Terraform)

---

## Step 3: Run Terraform to Provision the Infrastructure

1. Open terminal, navigate to the terraform directory:
   ```bash
   cd deploy/terraform
   ```
2. Initialize Terraform (downloads AWS providers):
   ```bash
   terraform init
   ```
3. Run Terraform Apply (this will read your public key and boot the EC2 instance with 4GB Swap enabled):
   ```powershell
   terraform apply -var "ssh_public_key=$(Get-Content -Raw "$HOME/.ssh/limiter_deploy.pub")" -auto-approve
   ```
4. Wait for it to complete. At the very end, it will output:
   `public_ip = "13.xxx.xxx.xxx"` (Copy this IP address!)

---

## Step 4: Configure GitHub Secrets
Go to your GitHub repository -> click **Settings** -> **Secrets and variables** -> **Actions** -> click **New repository secret**.

Create the following **four** secrets:

1. **`EC2_HOST`**:
   - Value: `[Your EC2 Public IP from step 3]`
2. **`EC2_SSH_KEY`**:
   - Value: `[Copy the exact contents of your private key file: C:\Users\R.K Singh\.ssh\limiter_deploy]` (Open it in Notepad to copy it).
3. **`NEXT_PUBLIC_API_URL`**:
   - Value: `http://[Your EC2 Public IP]/api/v1`
4. **`PROD_ENV_FILE`**:
   - Value: Paste the following production variables template (generate random values where indicated):
     ```env
     ENV=production
     PORT=8080

     # In-cluster Database details (auto-configured by docker compose)
     DB_USER=postgres
     DB_PASSWORD=limiterprodpassword123
     DB_NAME=ratelimiter

     # JWT Secrets (generate strong random hex strings)
     JWT_SECRET=8f9c1db2a3e4f5068a9b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f
     JWT_ACCESS_TTL=15m
     JWT_REFRESH_TTL=168h

     # Default admin credentials
     ADMIN_EMAIL=admin@ratelimiter.io
     ADMIN_PASSWORD=change-me-to-a-strong-password

     # Kafka Topic Configuration
     KAFKA_TOPIC=api_logs
     KAFKA_GROUP_ID=analytics_consumers

     # Public URLs for CORS & redirect flow
     APP_BASE_URL=http://[Your EC2 Public IP]
     CORS_ALLOWED_ORIGINS=http://[Your EC2 Public IP]
     ```

---

## Step 5: Push to GitHub to Auto-Deploy
Now push your changes to the `main` branch to trigger the active pipeline:
```bash
git add .
git commit -m "setup production deployment pipelines and build fixes"
git push origin main
```
Go to the **Actions** tab on your GitHub repository page. You will see the **Deploy to EC2** workflow running. Once it completes successfully (usually takes ~2-3 mins):
- Your application will be live at: `http://[Your EC2 Public IP]/`
- Your API will be live at: `http://[Your EC2 Public IP]/healthz`
- To log in to console, use: `admin@ratelimiter.io` with your admin password.
