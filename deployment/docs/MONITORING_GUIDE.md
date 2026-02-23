# ERP SPPG Monitoring & Alerting Guide

## Overview

Panduan ini menjelaskan sistem monitoring dan alerting untuk ERP SPPG, termasuk metrics yang dipantau, dashboard yang tersedia, dan prosedur troubleshooting.

## Monitoring Architecture

```
Application → Prometheus → Grafana → Alerts → Notification Channels
     ↓
Cloud Monitoring → Stackdriver → Alert Policies → Email/SMS
     ↓
Application Logs → Cloud Logging → Log-based Metrics → Alerts
```

## Key Metrics

### 1. Application Metrics

#### API Performance
- **Request Rate**: `rate(http_requests_total[5m])`
- **Response Time**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
- **Error Rate**: `rate(http_requests_total{status=~"5.."}[5m])`
- **Success Rate**: `rate(http_requests_total{status=~"2.."}[5m])`

#### Business Metrics
- **Active Users**: `active_users_total`
- **Login Attempts**: `auth_login_attempts_total`
- **Failed Logins**: `auth_login_failures_total`
- **Recipe Creations**: `recipes_created_total`
- **Menu Plans**: `menu_plans_created_total`
- **Purchase Orders**: `purchase_orders_created_total`
- **Deliveries Completed**: `deliveries_completed_total`

#### System Health
- **Database Connections**: `database_connections_active`
- **Redis Connections**: `redis_connections_active`
- **Firebase Connection**: `firebase_connection_status`
- **Background Jobs**: `background_jobs_pending`

### 2. Infrastructure Metrics

#### Server Resources
- **CPU Usage**: `100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
- **Memory Usage**: `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`
- **Disk Usage**: `(1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100`
- **Network I/O**: `rate(node_network_receive_bytes_total[5m])`

#### Database Metrics
- **Connection Count**: `pg_stat_database_numbackends`
- **Query Duration**: `pg_stat_statements_mean_time`
- **Lock Waits**: `pg_locks_count`
- **Replication Lag**: `pg_replication_lag_seconds`

#### Load Balancer Metrics
- **Request Count**: `nginx_http_requests_total`
- **Response Codes**: `nginx_http_requests_total{status}`
- **Upstream Response Time**: `nginx_upstream_response_time`
- **Active Connections**: `nginx_connections_active`

## Dashboards

### 1. Executive Dashboard
**URL**: `https://grafana.erp-sppg.example.com/d/executive`

**Panels**:
- System Health Overview
- Daily Active Users
- Business Metrics Summary
- Revenue/Cost Tracking
- SLA Compliance

### 2. Operations Dashboard
**URL**: `https://grafana.erp-sppg.example.com/d/operations`

**Panels**:
- API Response Times
- Error Rates by Endpoint
- Database Performance
- Server Resource Usage
- Alert Status

### 3. Business Dashboard
**URL**: `https://grafana.erp-sppg.example.com/d/business`

**Panels**:
- Recipe Management Activity
- Menu Planning Metrics
- Supply Chain KPIs
- Delivery Performance
- User Activity Patterns

### 4. Security Dashboard
**URL**: `https://grafana.erp-sppg.example.com/d/security`

**Panels**:
- Failed Login Attempts
- Suspicious Activity
- API Rate Limiting
- Security Events
- Access Patterns

## Alert Rules

### 1. Critical Alerts (Immediate Response)

#### Service Down
```yaml
alert: ServiceDown
expr: up == 0
for: 2m
severity: critical
description: "{{ $labels.job }} service is down"
```

#### High Error Rate
```yaml
alert: HighErrorRate
expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
for: 5m
severity: critical
description: "Error rate is {{ $value }} errors per second"
```

#### Database Connection Failure
```yaml
alert: DatabaseConnectionFailure
expr: up{job="postgres"} == 0
for: 2m
severity: critical
description: "PostgreSQL database is not responding"
```

#### SSL Certificate Expiring
```yaml
alert: SSLCertificateExpiring
expr: probe_ssl_earliest_cert_expiry - time() < 86400 * 7
for: 1h
severity: critical
description: "SSL certificate expires in {{ $value | humanizeDuration }}"
```

### 2. Warning Alerts (Monitor Closely)

#### High Response Time
```yaml
alert: HighResponseTime
expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
for: 5m
severity: warning
description: "95th percentile response time is {{ $value }} seconds"
```

#### High CPU Usage
```yaml
alert: HighCPUUsage
expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
for: 10m
severity: warning
description: "CPU usage is {{ $value }}% on {{ $labels.instance }}"
```

#### Low Disk Space
```yaml
alert: LowDiskSpace
expr: (1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100 > 85
for: 5m
severity: warning
description: "Disk usage is {{ $value }}% on {{ $labels.instance }}"
```

### 3. Business Alerts

#### Backup Failure
```yaml
alert: BackupFailure
expr: time() - backup_last_success_timestamp > 86400 * 2
for: 1h
severity: critical
description: "Database backup has not succeeded for more than 2 days"
```

#### High Failed Login Rate
```yaml
alert: HighFailedLoginRate
expr: rate(auth_login_failures_total[5m]) > 0.1
for: 5m
severity: warning
description: "Failed login rate is {{ $value }} per second"
```

#### Inventory Alerts Not Processed
```yaml
alert: InventoryAlertsNotProcessed
expr: inventory_low_stock_alerts_pending > 10
for: 10m
severity: warning
description: "{{ $value }} low stock alerts are pending processing"
```

## Notification Channels

### 1. Email Notifications
- **Critical Alerts**: ops-team@company.com, on-call@company.com
- **Warning Alerts**: ops-team@company.com
- **Business Alerts**: business-team@company.com

### 2. Slack Integration
- **Channel**: #erp-sppg-alerts
- **Critical**: @channel mention
- **Warning**: Regular message
- **Business**: #erp-sppg-business channel

### 3. SMS Alerts (Critical Only)
- On-call engineer
- Technical lead
- System administrator

### 4. PagerDuty Integration
- **Service**: ERP SPPG Production
- **Escalation**: L1 → L2 → L3 → Management
- **Schedule**: 24/7 coverage

## Log Management

### 1. Application Logs

#### Log Levels
- **ERROR**: Application errors, exceptions
- **WARN**: Performance issues, deprecated features
- **INFO**: Business events, user actions
- **DEBUG**: Detailed debugging information (dev only)

#### Log Format
```json
{
  "timestamp": "2024-12-01T10:30:00Z",
  "level": "INFO",
  "service": "erp-sppg-backend",
  "request_id": "req-123456",
  "user_id": "user-789",
  "message": "User logged in successfully",
  "metadata": {
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0..."
  }
}
```

### 2. Log Aggregation

#### Cloud Logging
- **Retention**: 30 days
- **Export**: BigQuery for long-term analysis
- **Filters**: Error logs, security events, business events

#### Log-based Metrics
- Error rate by service
- User activity patterns
- API usage statistics
- Security events

### 3. Log Analysis

#### Common Queries

**Error Analysis**:
```sql
SELECT 
  timestamp,
  jsonPayload.message,
  jsonPayload.user_id,
  jsonPayload.request_id
FROM `project.logs.erp_sppg_logs`
WHERE jsonPayload.level = "ERROR"
  AND timestamp >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 HOUR)
ORDER BY timestamp DESC
```

**User Activity**:
```sql
SELECT 
  jsonPayload.user_id,
  COUNT(*) as request_count
FROM `project.logs.erp_sppg_logs`
WHERE jsonPayload.level = "INFO"
  AND timestamp >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 DAY)
GROUP BY jsonPayload.user_id
ORDER BY request_count DESC
```

## Performance Monitoring

### 1. Application Performance Monitoring (APM)

#### Key Metrics
- **Apdex Score**: User satisfaction metric
- **Throughput**: Requests per minute
- **Response Time**: Average, 95th, 99th percentile
- **Error Rate**: Percentage of failed requests

#### Distributed Tracing
- Request flow across services
- Database query performance
- External API call latency
- Cache hit/miss rates

### 2. Real User Monitoring (RUM)

#### Frontend Metrics
- **Page Load Time**: Time to interactive
- **First Contentful Paint**: Visual loading metric
- **Cumulative Layout Shift**: Visual stability
- **JavaScript Errors**: Client-side errors

#### User Experience
- **Session Duration**: Average session length
- **Bounce Rate**: Single-page sessions
- **Feature Usage**: Most/least used features
- **Error Impact**: Users affected by errors

## Troubleshooting Runbooks

### 1. High Response Time

**Symptoms**: API responses > 1 second
**Investigation Steps**:
1. Check database query performance
2. Verify Redis cache hit rate
3. Check server resource usage
4. Review recent deployments
5. Analyze slow query logs

**Resolution**:
```bash
# Check database performance
docker exec -it erp-sppg-backend psql -c "
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;"

# Check Redis performance
docker exec -it erp-sppg-redis redis-cli info stats

# Scale up if needed
gcloud compute instance-groups managed resize erp-sppg-group --size=4
```

### 2. Database Connection Issues

**Symptoms**: Connection pool exhausted, timeouts
**Investigation Steps**:
1. Check active connections
2. Review connection pool settings
3. Identify long-running queries
4. Check for deadlocks

**Resolution**:
```bash
# Check active connections
docker exec -it erp-sppg-backend psql -c "
SELECT state, count(*) 
FROM pg_stat_activity 
GROUP BY state;"

# Kill long-running queries
docker exec -it erp-sppg-backend psql -c "
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'active' 
  AND query_start < now() - interval '5 minutes';"
```

### 3. High Memory Usage

**Symptoms**: Memory usage > 85%
**Investigation Steps**:
1. Check application memory usage
2. Review garbage collection metrics
3. Identify memory leaks
4. Check for large objects in cache

**Resolution**:
```bash
# Check container memory usage
docker stats

# Restart services if needed
docker-compose restart backend-1

# Scale up instance if persistent
gcloud compute instances set-machine-type erp-sppg-1 --machine-type=e2-standard-4
```

## Capacity Planning

### 1. Growth Projections

#### User Growth
- Current: 100 active users
- 6 months: 250 users
- 12 months: 500 users

#### Data Growth
- Database: 10GB → 50GB (12 months)
- File storage: 100GB → 1TB (12 months)
- Logs: 1GB/day → 5GB/day (12 months)

### 2. Resource Planning

#### Compute Resources
- **Current**: 2x e2-standard-2 instances
- **6 months**: 3x e2-standard-2 instances
- **12 months**: 4x e2-standard-4 instances

#### Database Resources
- **Current**: db-custom-2-4096
- **6 months**: db-custom-4-8192
- **12 months**: db-custom-8-16384

#### Storage Resources
- **Current**: 100GB SSD
- **6 months**: 500GB SSD
- **12 months**: 1TB SSD

### 3. Scaling Triggers

#### Automatic Scaling
- CPU usage > 70% for 5 minutes
- Memory usage > 80% for 5 minutes
- Response time > 500ms for 10 minutes

#### Manual Scaling
- Database connections > 80% of max
- Disk usage > 80%
- Network bandwidth > 80%

## Security Monitoring

### 1. Security Events

#### Authentication Events
- Failed login attempts
- Brute force attacks
- Unusual login patterns
- Privilege escalation attempts

#### Application Security
- SQL injection attempts
- XSS attempts
- CSRF attacks
- File upload abuse

#### Infrastructure Security
- Unauthorized access attempts
- Port scanning
- DDoS attacks
- Certificate issues

### 2. Security Alerts

#### Critical Security Alerts
- Multiple failed logins from same IP
- Admin account compromise
- Unusual data access patterns
- Security policy violations

#### Security Monitoring Tools
- Cloud Security Command Center
- Web Application Firewall logs
- Intrusion Detection System
- Vulnerability scanners

## Compliance & Reporting

### 1. SLA Reporting

#### Availability SLA: 99.9%
- **Monthly uptime**: Target > 99.9%
- **Response time**: 95th percentile < 500ms
- **Error rate**: < 0.1%

#### Performance SLA
- **API response time**: < 200ms average
- **Database query time**: < 100ms average
- **File upload time**: < 5 seconds

### 2. Compliance Reports

#### Monthly Reports
- System availability
- Performance metrics
- Security incidents
- Backup verification
- Capacity utilization

#### Quarterly Reports
- Trend analysis
- Capacity planning
- Security assessment
- Cost optimization
- Technology roadmap

## Contact Information

### Monitoring Team
- **Primary**: monitoring@company.com
- **Secondary**: ops-team@company.com
- **Emergency**: +62-xxx-xxx-xxxx

### Escalation Matrix
1. **L1 Support**: monitoring@company.com
2. **L2 Support**: senior-ops@company.com
3. **L3 Support**: tech-lead@company.com
4. **Management**: it-manager@company.com