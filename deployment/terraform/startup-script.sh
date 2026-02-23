#!/bin/bash

# ERP SPPG Startup Script for Compute Engine instances
# This script runs when instances start up

set -e

# Update system
apt-get update
apt-get install -y docker.io docker-compose curl

# Start Docker service
systemctl start docker
systemctl enable docker

# Add user to docker group
usermod -aG docker $USER

# Create application directory
mkdir -p /opt/erp-sppg
cd /opt/erp-sppg

# Download docker-compose configuration
gsutil cp gs://erp-sppg-config-${PROJECT_ID}/docker-compose.prod.yml .
gsutil cp gs://erp-sppg-config-${PROJECT_ID}/.env .

# Pull latest images
docker-compose -f docker-compose.prod.yml pull

# Start services
docker-compose -f docker-compose.prod.yml up -d

# Setup log rotation
cat > /etc/logrotate.d/erp-sppg << EOF
/opt/erp-sppg/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 root root
    postrotate
        docker-compose -f /opt/erp-sppg/docker-compose.prod.yml restart nginx
    endscript
}
EOF

# Setup monitoring agent
curl -sSO https://dl.google.com/cloudagents/add-google-cloud-ops-agent-repo.sh
bash add-google-cloud-ops-agent-repo.sh --also-install

# Configure monitoring
cat > /etc/google-cloud-ops-agent/config.yaml << EOF
logging:
  receivers:
    erp_sppg_logs:
      type: files
      include_paths:
        - /opt/erp-sppg/logs/*.log
      exclude_paths:
        - /opt/erp-sppg/logs/*.gz
  processors:
    erp_sppg_parser:
      type: parse_json
      field: message
  service:
    pipelines:
      default_pipeline:
        receivers: [erp_sppg_logs]
        processors: [erp_sppg_parser]

metrics:
  receivers:
    hostmetrics:
      type: hostmetrics
      collection_interval: 60s
  service:
    pipelines:
      default_pipeline:
        receivers: [hostmetrics]
EOF

# Restart ops agent
systemctl restart google-cloud-ops-agent

# Health check endpoint
cat > /opt/erp-sppg/health-check.sh << 'EOF'
#!/bin/bash
# Simple health check script

# Check if containers are running
if ! docker-compose -f /opt/erp-sppg/docker-compose.prod.yml ps | grep -q "Up"; then
    echo "ERROR: Some containers are not running"
    exit 1
fi

# Check if nginx is responding
if ! curl -f http://localhost/health > /dev/null 2>&1; then
    echo "ERROR: Nginx health check failed"
    exit 1
fi

echo "OK: All services healthy"
exit 0
EOF

chmod +x /opt/erp-sppg/health-check.sh

# Setup cron job for health checks
echo "*/5 * * * * root /opt/erp-sppg/health-check.sh >> /var/log/health-check.log 2>&1" >> /etc/crontab

# Setup automatic updates
cat > /opt/erp-sppg/update.sh << 'EOF'
#!/bin/bash
# Automatic update script

cd /opt/erp-sppg

# Pull latest images
docker-compose -f docker-compose.prod.yml pull

# Restart services with zero downtime
docker-compose -f docker-compose.prod.yml up -d --no-deps --build

# Clean up old images
docker image prune -f
EOF

chmod +x /opt/erp-sppg/update.sh

# Setup log forwarding to Cloud Logging
systemctl restart google-cloud-ops-agent

echo "ERP SPPG startup script completed successfully"