# ERP SPPG Deployment Guide

## Overview

Panduan ini menjelaskan cara melakukan deployment sistem ERP SPPG ke lingkungan produksi menggunakan Google Cloud Platform (GCP) dengan arsitektur High Availability.

## Arsitektur Sistem

```
Internet → Load Balancer → Nginx → Backend API (2 instances) → Cloud SQL (HA)
                      ↓
                   CDN → Static Files
                      ↓
                   Cloud Storage → File Uploads & Backups
                      ↓
                   Firebase → Real-time Updates
                      ↓
                   Redis → Cache & Sessions
```

## Prerequisites

### 1. Google Cloud Platform Setup
- GCP Project dengan billing enabled
- Service Account dengan permissions:
  - Compute Engine Admin
  - Cloud SQL Admin
  - Storage Admin
  - Monitoring Admin
  - Logging Admin

### 2. Tools Required
- Terraform >= 1.0
- Docker & Docker Compose
- gcloud CLI
- kubectl (jika menggunakan GKE)

### 3. Domain & SSL
- Domain name (contoh: erp-sppg.example.com)
- DNS management access

## Step 1: Infrastructure Deployment

### 1.1 Setup Terraform

```bash
cd deployment/terraform

# Initialize Terraform
terraform init

# Create terraform.tfvars
cat > terraform.tfvars << EOF
project_id = "your-gcp-project-id"
region     = "asia-southeast2"
zone       = "asia-southeast2-a"
environment = "prod"
EOF

# Plan deployment
terraform plan

# Apply infrastructure
terraform apply
```

### 1.2 Verify Infrastructure

```bash
# Check Cloud SQL instance
gcloud sql instances list

# Check Compute Engine instances
gcloud compute instances list

# Check Load Balancer
gcloud compute forwarding-rules list
```

## Step 2: Database Setup

### 2.1 Database Migration

```bash
# Connect to Cloud SQL instance
gcloud sql connect erp-sppg-db-prod --user=erp_sppg_user

# Run migrations (from backend directory)
cd backend
go run cmd/migrate/main.go up

# Verify tables created
\dt
```

### 2.2 Initial Data Setup

```bash
# Create initial admin user
go run cmd/seed/main.go --admin-user

# Import master data (if available)
go run cmd/seed/main.go --master-data
```

## Step 3: Application Deployment

### 3.1 Build Docker Images

```bash
# Build backend image
cd backend
docker build -t gcr.io/your-project-id/erp-sppg-backend:latest .
docker push gcr.io/your-project-id/erp-sppg-backend:latest

# Build web app image
cd ../web
docker build -t gcr.io/your-project-id/erp-sppg-web:latest .
docker push gcr.io/your-project-id/erp-sppg-web:latest

# Build PWA image
cd ../pwa
docker build -t gcr.io/your-project-id/erp-sppg-pwa:latest .
docker push gcr.io/your-project-id/erp-sppg-pwa:latest
```

### 3.2 Deploy to Compute Engine

```bash
# Copy deployment files to instances
gcloud compute scp deployment/docker-compose.prod.yml erp-sppg-1:~/
gcloud compute scp deployment/.env erp-sppg-1:~/

# SSH to instance and deploy
gcloud compute ssh erp-sppg-1

# On the instance:
sudo docker-compose -f docker-compose.prod.yml up -d
```

## Step 4: SSL Certificate & Domain Setup

### 4.1 DNS Configuration

```bash
# Get Load Balancer IP
gcloud compute forwarding-rules describe erp-sppg-forwarding-rule --global

# Configure DNS A record:
# erp-sppg.example.com → Load Balancer IP
```

### 4.2 SSL Certificate

SSL certificate akan otomatis di-provision oleh Google Managed SSL Certificate setelah DNS dikonfigurasi.

```bash
# Check SSL certificate status
gcloud compute ssl-certificates describe erp-sppg-ssl-cert --global
```

## Step 5: Monitoring Setup

### 5.1 Prometheus & Grafana

```bash
# Access Grafana dashboard
# URL: http://your-instance-ip:3000
# Username: admin
# Password: (from .env file)

# Import ERP SPPG dashboard
# File: deployment/monitoring/grafana/dashboards/erp-sppg-dashboard.json
```

### 5.2 Cloud Monitoring

```bash
# Enable Cloud Monitoring alerts
gcloud alpha monitoring policies create --policy-from-file=monitoring/alert-policies.yaml
```

## Step 6: Backup Configuration

### 6.1 Automated Database Backup

```bash
# Backup script sudah dikonfigurasi di docker-compose
# Backup berjalan setiap hari jam 3 pagi
# Retention: 30 hari

# Manual backup
docker exec erp-sppg-backup /backup.sh
```

### 6.2 File Backup

```bash
# Setup Cloud Storage lifecycle
gsutil lifecycle set deployment/storage-lifecycle.json gs://erp-sppg-backups-your-project-id
```

## Step 7: Security Hardening

### 7.1 Firewall Rules

```bash
# Firewall rules sudah dikonfigurasi via Terraform
# Hanya port 80, 443 yang terbuka untuk public

# Verify firewall rules
gcloud compute firewall-rules list
```

### 7.2 IAM & Service Accounts

```bash
# Service account sudah dikonfigurasi dengan minimal permissions
# Verify IAM bindings
gcloud projects get-iam-policy your-project-id
```

## Step 8: Testing & Validation

### 8.1 Health Checks

```bash
# Test application endpoints
curl https://erp-sppg.example.com/health
curl https://erp-sppg.example.com/api/v1/health

# Test authentication
curl -X POST https://erp-sppg.example.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'
```

### 8.2 Load Testing

```bash
# Install Apache Bench
sudo apt-get install apache2-utils

# Basic load test
ab -n 1000 -c 10 https://erp-sppg.example.com/api/v1/health
```

## Step 9: Go Live Checklist

### Pre-Launch
- [ ] Infrastructure deployed dan tested
- [ ] Database migrated dan seeded
- [ ] SSL certificate active
- [ ] Monitoring configured
- [ ] Backup tested
- [ ] Load testing completed
- [ ] Security scan completed
- [ ] User acceptance testing completed

### Launch
- [ ] DNS switched to production
- [ ] Application accessible
- [ ] All features working
- [ ] Real-time updates working
- [ ] File uploads working
- [ ] Notifications working

### Post-Launch
- [ ] Monitor system metrics
- [ ] Check error logs
- [ ] Verify backup completion
- [ ] User feedback collection
- [ ] Performance optimization

## Troubleshooting

### Common Issues

#### 1. Database Connection Issues
```bash
# Check Cloud SQL instance status
gcloud sql instances describe erp-sppg-db-prod

# Check private IP connectivity
gcloud compute ssh erp-sppg-1 --command="ping CLOUD_SQL_PRIVATE_IP"
```

#### 2. SSL Certificate Issues
```bash
# Check certificate status
gcloud compute ssl-certificates describe erp-sppg-ssl-cert --global

# Common causes:
# - DNS not pointing to Load Balancer IP
# - Domain verification pending
```

#### 3. Application Not Starting
```bash
# Check container logs
docker-compose -f docker-compose.prod.yml logs backend-1

# Check environment variables
docker-compose -f docker-compose.prod.yml config
```

#### 4. High Response Times
```bash
# Check database performance
# Check Redis connectivity
# Check resource utilization
# Scale up instances if needed
```

## Maintenance

### Regular Tasks
- Monitor system metrics daily
- Review error logs weekly
- Update dependencies monthly
- Security patches as needed
- Backup verification weekly

### Scaling
- Monitor CPU/Memory usage
- Scale Compute Engine instances as needed
- Upgrade Cloud SQL instance if database becomes bottleneck
- Implement read replicas for heavy read workloads

## Support Contacts

- **Infrastructure**: DevOps Team
- **Application**: Development Team
- **Database**: DBA Team
- **Security**: Security Team

## References

- [Google Cloud SQL Documentation](https://cloud.google.com/sql/docs)
- [Google Cloud Load Balancing](https://cloud.google.com/load-balancing/docs)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Terraform GCP Provider](https://registry.terraform.io/providers/hashicorp/google/latest/docs)