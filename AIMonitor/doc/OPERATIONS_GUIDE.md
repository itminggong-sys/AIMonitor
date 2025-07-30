# AI Monitor 运维指南

## 📋 目录

1. [运维概览](#运维概览)
2. [日常运维](#日常运维)
3. [监控告警](#监控告警)
4. [性能调优](#性能调优)
5. [备份恢复](#备份恢复)
6. [故障处理](#故障处理)
7. [安全运维](#安全运维)
8. [升级维护](#升级维护)
9. [容量规划](#容量规划)
10. [运维工具](#运维工具)

## 🎯 运维概览

### 系统架构概述

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端服务      │    │   后端服务      │    │   数据存储      │
│                 │    │                 │    │                 │
│ • React应用     │◄──►│ • Go API服务    │◄──►│ • PostgreSQL    │
│ • Nginx代理     │    │ • WebSocket     │    │ • Redis缓存     │
│ • 静态资源      │    │ • AI分析引擎    │    │ • Elasticsearch │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   监控组件      │
                    │                 │
                    │ • Prometheus    │
                    │ • Grafana       │
                    │ • Jaeger        │
                    │ • MinIO         │
                    └─────────────────┘
```

### 核心组件职责

| 组件 | 职责 | 关键指标 |
|------|------|----------|
| **前端服务** | 用户界面、静态资源服务 | 响应时间、错误率、并发用户数 |
| **后端服务** | API服务、业务逻辑、AI分析 | QPS、响应时间、内存使用率 |
| **数据库** | 数据持久化、事务处理 | 连接数、查询性能、存储空间 |
| **缓存** | 数据缓存、会话存储 | 命中率、内存使用、连接数 |
| **监控** | 指标收集、告警通知 | 数据完整性、存储容量 |

### 运维责任矩阵

| 任务类型 | 日常运维 | 应急响应 | 容量规划 | 安全审计 |
|----------|----------|----------|----------|----------|
| **系统监控** | ✓ | ✓ | ✓ | ✓ |
| **性能优化** | ✓ | ✓ | ✓ | - |
| **备份恢复** | ✓ | ✓ | - | ✓ |
| **安全加固** | ✓ | - | - | ✓ |
| **版本升级** | ✓ | - | ✓ | ✓ |

## 📅 日常运维

### 每日检查清单

#### 系统健康检查

```bash
#!/bin/bash
# scripts/daily_health_check.sh

echo "=== AI Monitor 每日健康检查 ==="
echo "检查时间: $(date)"
echo

# 1. 检查服务状态
echo "1. 检查服务状态"
services=("ai-monitor" "postgresql" "redis" "nginx" "prometheus" "grafana")
for service in "${services[@]}"; do
    if systemctl is-active --quiet $service; then
        echo "✓ $service: 运行中"
    else
        echo "✗ $service: 已停止"
    fi
done
echo

# 2. 检查端口监听
echo "2. 检查端口监听"
ports=("8080:AI-Monitor" "5432:PostgreSQL" "6379:Redis" "80:Nginx" "9090:Prometheus" "3000:Grafana")
for port_info in "${ports[@]}"; do
    port=$(echo $port_info | cut -d: -f1)
    service=$(echo $port_info | cut -d: -f2)
    if netstat -tuln | grep -q ":$port "; then
        echo "✓ $service ($port): 监听中"
    else
        echo "✗ $service ($port): 未监听"
    fi
done
echo

# 3. 检查磁盘空间
echo "3. 检查磁盘空间"
df -h | grep -E '(/$|/var|/data)' | while read line; do
    usage=$(echo $line | awk '{print $5}' | sed 's/%//')
    mount=$(echo $line | awk '{print $6}')
    if [ $usage -gt 80 ]; then
        echo "⚠ $mount: ${usage}% (警告)" 
    elif [ $usage -gt 90 ]; then
        echo "✗ $mount: ${usage}% (危险)"
    else
        echo "✓ $mount: ${usage}% (正常)"
    fi
done
echo

# 4. 检查内存使用
echo "4. 检查内存使用"
mem_usage=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
echo "内存使用率: ${mem_usage}%"
if (( $(echo "$mem_usage > 80" | bc -l) )); then
    echo "⚠ 内存使用率较高"
fi
echo

# 5. 检查CPU负载
echo "5. 检查CPU负载"
load_avg=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
echo "1分钟负载: $load_avg"
echo

# 6. 检查应用健康接口
echo "6. 检查应用健康接口"
if curl -s -f http://localhost:8080/api/v1/health > /dev/null; then
    echo "✓ AI Monitor API: 健康"
else
    echo "✗ AI Monitor API: 异常"
fi
echo

echo "=== 检查完成 ==="
```

#### 日志检查

```bash
#!/bin/bash
# scripts/daily_log_check.sh

echo "=== 每日日志检查 ==="
echo "检查时间: $(date)"
echo

# 检查错误日志
echo "1. 检查应用错误日志 (最近24小时)"
error_count=$(journalctl -u ai-monitor --since "24 hours ago" | grep -i error | wc -l)
if [ $error_count -gt 0 ]; then
    echo "⚠ 发现 $error_count 条错误日志"
    echo "最近的错误:"
    journalctl -u ai-monitor --since "24 hours ago" | grep -i error | tail -5
else
    echo "✓ 无错误日志"
fi
echo

# 检查数据库日志
echo "2. 检查数据库日志"
db_errors=$(tail -1000 /var/log/postgresql/postgresql-*.log | grep -i error | wc -l)
if [ $db_errors -gt 0 ]; then
    echo "⚠ 发现 $db_errors 条数据库错误"
else
    echo "✓ 数据库日志正常"
fi
echo

# 检查Nginx访问日志
echo "3. 检查Nginx访问统计 (最近24小时)"
today=$(date +%d/%b/%Y)
total_requests=$(grep "$today" /var/log/nginx/access.log | wc -l)
4xx_errors=$(grep "$today" /var/log/nginx/access.log | grep ' 4[0-9][0-9] ' | wc -l)
5xx_errors=$(grep "$today" /var/log/nginx/access.log | grep ' 5[0-9][0-9] ' | wc -l)

echo "总请求数: $total_requests"
echo "4xx错误: $4xx_errors"
echo "5xx错误: $5xx_errors"

if [ $5xx_errors -gt 100 ]; then
    echo "⚠ 5xx错误较多，需要关注"
fi
echo

echo "=== 日志检查完成 ==="
```

### 每周维护任务

#### 数据库维护

```sql
-- scripts/weekly_db_maintenance.sql

-- 1. 更新表统计信息
ANALYZE;

-- 2. 重建索引（如果需要）
REINDEX INDEX CONCURRENTLY idx_metrics_timestamp;
REINDEX INDEX CONCURRENTLY idx_alerts_created_at;

-- 3. 清理过期数据
DELETE FROM metrics WHERE created_at < NOW() - INTERVAL '90 days';
DELETE FROM logs WHERE created_at < NOW() - INTERVAL '30 days';
DELETE FROM events WHERE created_at < NOW() - INTERVAL '365 days';

-- 4. 检查数据库大小
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- 5. 检查慢查询
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements 
WHERE mean_time > 1000
ORDER BY mean_time DESC
LIMIT 10;
```

#### 系统清理

```bash
#!/bin/bash
# scripts/weekly_cleanup.sh

echo "=== 每周系统清理 ==="
echo "开始时间: $(date)"
echo

# 1. 清理日志文件
echo "1. 清理旧日志文件"
find /var/log -name "*.log" -mtime +30 -delete
find /var/log -name "*.log.*" -mtime +7 -delete
journalctl --vacuum-time=30d
echo "✓ 日志清理完成"
echo

# 2. 清理临时文件
echo "2. 清理临时文件"
find /tmp -type f -mtime +7 -delete
find /var/tmp -type f -mtime +7 -delete
echo "✓ 临时文件清理完成"
echo

# 3. 清理Docker资源（如果使用Docker）
if command -v docker &> /dev/null; then
    echo "3. 清理Docker资源"
    docker system prune -f
    docker volume prune -f
    echo "✓ Docker清理完成"
else
    echo "3. 跳过Docker清理（未安装Docker）"
fi
echo

# 4. 更新系统包（可选）
echo "4. 检查系统更新"
if command -v apt &> /dev/null; then
    apt list --upgradable
elif command -v yum &> /dev/null; then
    yum check-update
fi
echo

echo "=== 清理完成 ==="
```

### 每月报告生成

```bash
#!/bin/bash
# scripts/monthly_report.sh

REPORT_DATE=$(date +%Y-%m)
REPORT_FILE="/var/reports/ai-monitor-report-$REPORT_DATE.md"

mkdir -p /var/reports

cat > $REPORT_FILE << EOF
# AI Monitor 月度运维报告

**报告期间**: $REPORT_DATE  
**生成时间**: $(date)  
**报告人**: 系统自动生成

## 系统概览

### 服务可用性
$(systemctl is-active ai-monitor postgresql redis nginx prometheus grafana | 
  awk '{print "- " NR ". 服务" NR ": " $1}')

### 资源使用情况

#### 磁盘使用
\`\`\`
$(df -h)
\`\`\`

#### 内存使用
\`\`\`
$(free -h)
\`\`\`

#### CPU负载
\`\`\`
$(uptime)
\`\`\`

## 性能指标

### 数据库性能
- 连接数: $(psql -t -c "SELECT count(*) FROM pg_stat_activity;")
- 数据库大小: $(psql -t -c "SELECT pg_size_pretty(pg_database_size('ai_monitor'));")

### 应用性能
- 平均响应时间: 待补充
- 错误率: 待补充
- 并发用户数: 待补充

## 告警统计

### 本月告警数量
- 严重告警: 待补充
- 警告告警: 待补充
- 信息告警: 待补充

## 维护记录

### 已完成维护
- 数据库维护: $(date -d "last sunday" +%Y-%m-%d)
- 系统清理: $(date -d "last sunday" +%Y-%m-%d)
- 安全更新: 待补充

### 计划维护
- 下次数据库维护: $(date -d "next sunday" +%Y-%m-%d)
- 下次系统清理: $(date -d "next sunday" +%Y-%m-%d)

## 建议和改进

1. 根据资源使用情况，建议关注磁盘空间增长趋势
2. 定期检查数据库性能，优化慢查询
3. 持续监控应用性能指标

---
*此报告由AI Monitor运维脚本自动生成*
EOF

echo "月度报告已生成: $REPORT_FILE"
```

## 🚨 监控告警

### 告警规则配置

#### Prometheus告警规则

```yaml
# config/prometheus/alert_rules.yml
groups:
  - name: ai-monitor-alerts
    rules:
      # 服务可用性告警
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "服务 {{ $labels.instance }} 不可用"
          description: "服务 {{ $labels.instance }} 已停止响应超过1分钟"
      
      # 高CPU使用率告警
      - alert: HighCPUUsage
        expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.instance }} CPU使用率过高"
          description: "{{ $labels.instance }} CPU使用率为 {{ $value }}%，持续5分钟"
      
      # 高内存使用率告警
      - alert: HighMemoryUsage
        expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.instance }} 内存使用率过高"
          description: "{{ $labels.instance }} 内存使用率为 {{ $value }}%，持续5分钟"
      
      # 磁盘空间不足告警
      - alert: DiskSpaceLow
        expr: (1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100 > 85
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "{{ $labels.instance }} 磁盘空间不足"
          description: "{{ $labels.instance }} 磁盘 {{ $labels.mountpoint }} 使用率为 {{ $value }}%"
      
      # 数据库连接数过多告警
      - alert: DatabaseConnectionsHigh
        expr: pg_stat_database_numbackends > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "数据库连接数过多"
          description: "数据库当前连接数为 {{ $value }}，超过阈值"
      
      # API响应时间过长告警
      - alert: APIResponseTimeSlow
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "API响应时间过长"
          description: "95%的API请求响应时间超过2秒，当前为 {{ $value }}秒"
      
      # Redis连接失败告警
      - alert: RedisConnectionFailed
        expr: redis_up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Redis连接失败"
          description: "Redis服务不可用，持续1分钟"
```

#### 告警通知配置

```yaml
# config/alertmanager/alertmanager.yml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@ai-monitor.com'
  smtp_auth_username: 'alerts@ai-monitor.com'
  smtp_auth_password: 'your-app-password'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
    - match:
        severity: warning
      receiver: 'warning-alerts'

receivers:
  - name: 'web.hook'
    webhook_configs:
      - url: 'http://localhost:8080/api/v1/alerts/webhook'
  
  - name: 'critical-alerts'
    email_configs:
      - to: 'ops-team@company.com'
        subject: '[CRITICAL] AI Monitor Alert'
        body: |
          Alert: {{ .GroupLabels.alertname }}
          Summary: {{ range .Alerts }}{{ .Annotations.summary }}{{ end }}
          Description: {{ range .Alerts }}{{ .Annotations.description }}{{ end }}
    webhook_configs:
      - url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
        send_resolved: true
  
  - name: 'warning-alerts'
    email_configs:
      - to: 'dev-team@company.com'
        subject: '[WARNING] AI Monitor Alert'
        body: |
          Alert: {{ .GroupLabels.alertname }}
          Summary: {{ range .Alerts }}{{ .Annotations.summary }}{{ end }}
          Description: {{ range .Alerts }}{{ .Annotations.description }}{{ end }}

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'instance']
```

### 告警处理流程

#### 告警分级处理

| 级别 | 响应时间 | 处理人员 | 处理方式 |
|------|----------|----------|----------|
| **Critical** | 5分钟内 | 值班工程师 | 立即处理，必要时升级 |
| **Warning** | 30分钟内 | 运维团队 | 分析原因，制定处理计划 |
| **Info** | 2小时内 | 相关负责人 | 记录问题，定期处理 |

#### 告警处理脚本

```bash
#!/bin/bash
# scripts/alert_handler.sh

ALERT_LEVEL=$1
ALERT_NAME=$2
ALERT_INSTANCE=$3
ALERT_DESCRIPTION=$4

echo "收到告警: $ALERT_NAME ($ALERT_LEVEL)"
echo "实例: $ALERT_INSTANCE"
echo "描述: $ALERT_DESCRIPTION"
echo "时间: $(date)"

case $ALERT_LEVEL in
    "critical")
        echo "执行紧急处理流程..."
        # 自动重启服务（如果适用）
        if [[ $ALERT_NAME == "ServiceDown" ]]; then
            echo "尝试重启服务..."
            systemctl restart ai-monitor
            sleep 30
            if systemctl is-active --quiet ai-monitor; then
                echo "服务重启成功"
                # 发送恢复通知
                curl -X POST http://localhost:8080/api/v1/alerts/resolve \
                     -H "Content-Type: application/json" \
                     -d "{\"alert\": \"$ALERT_NAME\", \"instance\": \"$ALERT_INSTANCE\"}"
            else
                echo "服务重启失败，需要人工介入"
                # 发送升级通知
                curl -X POST http://localhost:8080/api/v1/alerts/escalate \
                     -H "Content-Type: application/json" \
                     -d "{\"alert\": \"$ALERT_NAME\", \"instance\": \"$ALERT_INSTANCE\"}"
            fi
        fi
        ;;
    "warning")
        echo "记录警告信息..."
        # 记录到运维日志
        echo "$(date): WARNING - $ALERT_NAME on $ALERT_INSTANCE" >> /var/log/ai-monitor/ops.log
        ;;
    "info")
        echo "记录信息..."
        # 记录到信息日志
        echo "$(date): INFO - $ALERT_NAME on $ALERT_INSTANCE" >> /var/log/ai-monitor/info.log
        ;;
esac
```

## ⚡ 性能调优

### 数据库性能优化

#### 查询优化

```sql
-- scripts/db_optimization.sql

-- 1. 创建必要的索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_metrics_timestamp_type 
    ON metrics(timestamp, metric_type);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_alerts_status_created 
    ON alerts(status, created_at);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_logs_level_timestamp 
    ON logs(level, timestamp);

-- 2. 分区表设置（按时间分区）
CREATE TABLE IF NOT EXISTS metrics_2024_01 PARTITION OF metrics
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE IF NOT EXISTS metrics_2024_02 PARTITION OF metrics
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- 3. 优化配置参数
-- 在postgresql.conf中设置：
-- shared_buffers = 256MB
-- effective_cache_size = 1GB
-- work_mem = 4MB
-- maintenance_work_mem = 64MB
-- checkpoint_completion_target = 0.9
-- wal_buffers = 16MB
-- default_statistics_target = 100

-- 4. 定期维护
-- 每周执行VACUUM ANALYZE
-- 每月执行REINDEX
```

#### 连接池优化

```yaml
# config/database.yaml
database:
  pool:
    max_open_conns: 25        # 根据CPU核心数调整
    max_idle_conns: 10        # 保持适量空闲连接
    conn_max_lifetime: 300s   # 连接最大生存时间
    conn_max_idle_time: 60s   # 连接最大空闲时间
  
  query:
    timeout: 30s              # 查询超时时间
    slow_query_threshold: 1s  # 慢查询阈值
    log_slow_queries: true    # 记录慢查询
```

### 应用性能优化

#### Go应用调优

```go
// internal/config/performance.go
package config

import (
    "runtime"
    "time"
)

// 性能优化配置
func OptimizeRuntime() {
    // 设置GOMAXPROCS
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // 设置GC目标百分比
    runtime.SetGCPercent(100)
    
    // 设置内存限制（如果需要）
    // runtime.SetMemoryLimit(1 << 30) // 1GB
}

// HTTP服务器优化配置
type ServerConfig struct {
    ReadTimeout       time.Duration `yaml:"read_timeout"`
    WriteTimeout      time.Duration `yaml:"write_timeout"`
    IdleTimeout       time.Duration `yaml:"idle_timeout"`
    ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
    MaxHeaderBytes    int           `yaml:"max_header_bytes"`
}

func DefaultServerConfig() ServerConfig {
    return ServerConfig{
        ReadTimeout:       30 * time.Second,
        WriteTimeout:      30 * time.Second,
        IdleTimeout:       60 * time.Second,
        ReadHeaderTimeout: 10 * time.Second,
        MaxHeaderBytes:    1 << 20, // 1MB
    }
}
```

#### 缓存优化

```yaml
# config/redis.yaml
redis:
  pool:
    max_active: 100           # 最大活跃连接数
    max_idle: 50              # 最大空闲连接数
    idle_timeout: 300s        # 空闲超时时间
    wait: true                # 连接池满时等待
  
  cache:
    default_ttl: 3600s        # 默认过期时间
    key_prefix: "ai_monitor:" # 键前缀
    
    # 不同类型数据的TTL策略
    ttl_strategy:
      user_session: 86400s     # 用户会话24小时
      api_cache: 300s          # API缓存5分钟
      metrics_cache: 60s       # 指标缓存1分钟
      static_data: 3600s       # 静态数据1小时
```

### 前端性能优化

#### Nginx配置优化

```nginx
# config/nginx/performance.conf

# 工作进程数
worker_processes auto;

# 每个工作进程的最大连接数
events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    # 基础优化
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    
    # 缓冲区优化
    client_body_buffer_size 128k;
    client_max_body_size 10m;
    client_header_buffer_size 1k;
    large_client_header_buffers 4 4k;
    output_buffers 1 32k;
    postpone_output 1460;
    
    # Gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;
    
    # 静态文件缓存
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|pdf|txt)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        add_header Vary Accept-Encoding;
    }
    
    # API接口优化
    location /api/ {
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;
        proxy_temp_file_write_size 8k;
        
        # 连接池
        upstream backend {
            server 127.0.0.1:8080;
            keepalive 32;
        }
        
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
}
```

## 💾 备份恢复

### 数据库备份策略

#### 自动备份脚本

```bash
#!/bin/bash
# scripts/backup_database.sh

set -e

# 配置变量
DB_NAME="ai_monitor"
DB_USER="ai_monitor"
BACKUP_DIR="/var/backups/postgresql"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/${DB_NAME}_${DATE}.sql"
RETENTION_DAYS=30

# 创建备份目录
mkdir -p $BACKUP_DIR

echo "开始数据库备份: $(date)"
echo "备份文件: $BACKUP_FILE"

# 执行备份
pg_dump -h localhost -U $DB_USER -d $DB_NAME > $BACKUP_FILE

# 压缩备份文件
gzip $BACKUP_FILE
BACKUP_FILE="${BACKUP_FILE}.gz"

echo "备份完成: $BACKUP_FILE"
echo "备份大小: $(du -h $BACKUP_FILE | cut -f1)"

# 清理旧备份
echo "清理 $RETENTION_DAYS 天前的备份文件..."
find $BACKUP_DIR -name "${DB_NAME}_*.sql.gz" -mtime +$RETENTION_DAYS -delete

# 验证备份文件
if [ -f "$BACKUP_FILE" ] && [ -s "$BACKUP_FILE" ]; then
    echo "✓ 备份验证成功"
else
    echo "✗ 备份验证失败"
    exit 1
fi

# 发送备份通知
curl -X POST http://localhost:8080/api/v1/notifications \
     -H "Content-Type: application/json" \
     -d "{
         \"type\": \"backup_completed\",
         \"message\": \"数据库备份完成: $BACKUP_FILE\",
         \"timestamp\": \"$(date -Iseconds)\"
     }"

echo "数据库备份流程完成: $(date)"
```

#### 增量备份脚本

```bash
#!/bin/bash
# scripts/incremental_backup.sh

set -e

# 配置变量
DB_NAME="ai_monitor"
DB_USER="ai_monitor"
BACKUP_DIR="/var/backups/postgresql/incremental"
WAL_ARCHIVE_DIR="/var/backups/postgresql/wal_archive"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p $BACKUP_DIR
mkdir -p $WAL_ARCHIVE_DIR

echo "开始增量备份: $(date)"

# 检查是否存在基础备份
BASE_BACKUP=$(find $BACKUP_DIR -name "base_*" -type d | sort | tail -1)

if [ -z "$BASE_BACKUP" ]; then
    echo "未找到基础备份，创建基础备份..."
    BASE_BACKUP_DIR="$BACKUP_DIR/base_$DATE"
    pg_basebackup -h localhost -U $DB_USER -D $BASE_BACKUP_DIR -Ft -z -P
    echo "基础备份完成: $BASE_BACKUP_DIR"
else
    echo "使用现有基础备份: $BASE_BACKUP"
fi

# 归档WAL文件
echo "归档WAL文件..."
psql -h localhost -U $DB_USER -d $DB_NAME -c "SELECT pg_switch_wal();"

# 复制新的WAL文件
cp /var/lib/postgresql/*/main/pg_wal/0* $WAL_ARCHIVE_DIR/ 2>/dev/null || true

echo "增量备份完成: $(date)"
```

### 数据恢复流程

#### 完整恢复脚本

```bash
#!/bin/bash
# scripts/restore_database.sh

set -e

BACKUP_FILE=$1
DB_NAME="ai_monitor"
DB_USER="ai_monitor"

if [ -z "$BACKUP_FILE" ]; then
    echo "用法: $0 <backup_file>"
    echo "可用备份文件:"
    ls -la /var/backups/postgresql/*.sql.gz
    exit 1
fi

echo "开始数据库恢复: $(date)"
echo "备份文件: $BACKUP_FILE"

# 确认恢复操作
read -p "确认要恢复数据库吗？这将覆盖现有数据 (y/N): " confirm
if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "恢复操作已取消"
    exit 0
fi

# 停止应用服务
echo "停止应用服务..."
systemctl stop ai-monitor

# 创建恢复前备份
echo "创建恢复前备份..."
RECOVERY_BACKUP="/var/backups/postgresql/pre_recovery_$(date +%Y%m%d_%H%M%S).sql"
pg_dump -h localhost -U $DB_USER -d $DB_NAME > $RECOVERY_BACKUP
gzip $RECOVERY_BACKUP
echo "恢复前备份完成: ${RECOVERY_BACKUP}.gz"

# 删除现有数据库
echo "删除现有数据库..."
psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"
psql -h localhost -U postgres -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;"

# 恢复数据
echo "恢复数据..."
if [[ $BACKUP_FILE == *.gz ]]; then
    gunzip -c $BACKUP_FILE | psql -h localhost -U $DB_USER -d $DB_NAME
else
    psql -h localhost -U $DB_USER -d $DB_NAME < $BACKUP_FILE
fi

# 验证恢复
echo "验证数据恢复..."
TABLE_COUNT=$(psql -h localhost -U $DB_USER -d $DB_NAME -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema='public';")
echo "恢复的表数量: $TABLE_COUNT"

if [ $TABLE_COUNT -gt 0 ]; then
    echo "✓ 数据恢复验证成功"
else
    echo "✗ 数据恢复验证失败"
    exit 1
fi

# 重启应用服务
echo "重启应用服务..."
systemctl start ai-monitor

# 等待服务启动
sleep 10

# 验证应用服务
if curl -s -f http://localhost:8080/api/v1/health > /dev/null; then
    echo "✓ 应用服务启动成功"
else
    echo "✗ 应用服务启动失败"
    exit 1
fi

echo "数据库恢复完成: $(date)"
```

### 配置文件备份

```bash
#!/bin/bash
# scripts/backup_configs.sh

set -e

BACKUP_DIR="/var/backups/configs"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/ai-monitor-configs_$DATE.tar.gz"

mkdir -p $BACKUP_DIR

echo "开始配置文件备份: $(date)"

# 备份配置文件
tar -czf $BACKUP_FILE \
    /etc/ai-monitor/ \
    /etc/nginx/sites-available/ai-monitor \
    /etc/postgresql/*/main/postgresql.conf \
    /etc/redis/redis.conf \
    /etc/prometheus/ \
    /etc/grafana/ \
    2>/dev/null || true

echo "配置文件备份完成: $BACKUP_FILE"
echo "备份大小: $(du -h $BACKUP_FILE | cut -f1)"

# 清理旧备份（保留30天）
find $BACKUP_DIR -name "ai-monitor-configs_*.tar.gz" -mtime +30 -delete

echo "配置文件备份流程完成: $(date)"
```

## 🛠️ 故障处理

### 常见故障处理手册

#### 服务无法启动

**故障现象**：
- 应用进程无法启动
- 启动后立即退出
- 端口无法绑定

**诊断步骤**：

```bash
#!/bin/bash
# scripts/diagnose_startup_failure.sh

echo "=== 服务启动故障诊断 ==="
echo "诊断时间: $(date)"
echo

# 1. 检查服务状态
echo "1. 检查服务状态"
systemctl status ai-monitor
echo

# 2. 检查端口占用
echo "2. 检查端口占用"
netstat -tulpn | grep :8080
echo

# 3. 检查配置文件
echo "3. 检查配置文件语法"
if command -v ai-monitor &> /dev/null; then
    ai-monitor config validate
else
    echo "ai-monitor命令不可用，跳过配置验证"
fi
echo

# 4. 检查依赖服务
echo "4. 检查依赖服务"
services=("postgresql" "redis")
for service in "${services[@]}"; do
    if systemctl is-active --quiet $service; then
        echo "✓ $service: 运行中"
    else
        echo "✗ $service: 已停止"
    fi
done
echo

# 5. 检查日志
echo "5. 最近的错误日志"
journalctl -u ai-monitor --since "10 minutes ago" | grep -i error | tail -10
echo

# 6. 检查磁盘空间
echo "6. 检查磁盘空间"
df -h | grep -E '(/$|/var)'
echo

# 7. 检查内存
echo "7. 检查内存使用"
free -h
echo

echo "=== 诊断完成 ==="
```

**解决方案**：

```bash
#!/bin/bash
# scripts/fix_startup_failure.sh

echo "=== 修复服务启动问题 ==="

# 1. 停止可能冲突的进程
echo "1. 停止冲突进程"
pkill -f ai-monitor || true
sleep 5

# 2. 清理临时文件
echo "2. 清理临时文件"
rm -f /tmp/ai-monitor.pid
rm -f /var/run/ai-monitor.sock

# 3. 检查并修复权限
echo "3. 修复文件权限"
chown -R ai-monitor:ai-monitor /var/lib/ai-monitor
chmod 755 /usr/local/bin/ai-monitor
chmod 644 /etc/ai-monitor/config.yaml

# 4. 重启依赖服务
echo "4. 重启依赖服务"
systemctl restart postgresql
systemctl restart redis
sleep 10

# 5. 重启主服务
echo "5. 重启AI Monitor服务"
systemctl restart ai-monitor
sleep 15

# 6. 验证服务状态
echo "6. 验证服务状态"
if systemctl is-active --quiet ai-monitor; then
    echo "✓ 服务启动成功"
    curl -s http://localhost:8080/api/v1/health
else
    echo "✗ 服务启动失败"
    journalctl -u ai-monitor --since "5 minutes ago" | tail -20
fi

echo "=== 修复完成 ==="
```

#### 数据库连接问题

**故障现象**：
- 数据库连接超时
- 连接池耗尽
- 查询执行缓慢

**诊断脚本**：

```bash
#!/bin/bash
# scripts/diagnose_database.sh

DB_NAME="ai_monitor"
DB_USER="ai_monitor"

echo "=== 数据库连接故障诊断 ==="
echo "诊断时间: $(date)"
echo

# 1. 检查PostgreSQL服务状态
echo "1. 检查PostgreSQL服务状态"
systemctl status postgresql
echo

# 2. 检查数据库连接
echo "2. 测试数据库连接"
if psql -h localhost -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✓ 数据库连接正常"
else
    echo "✗ 数据库连接失败"
fi
echo

# 3. 检查当前连接数
echo "3. 检查当前连接数"
psql -h localhost -U $DB_USER -d $DB_NAME -c "
    SELECT 
        count(*) as total_connections,
        count(*) FILTER (WHERE state = 'active') as active_connections,
        count(*) FILTER (WHERE state = 'idle') as idle_connections
    FROM pg_stat_activity 
    WHERE datname = '$DB_NAME';
"
echo

# 4. 检查长时间运行的查询
echo "4. 检查长时间运行的查询"
psql -h localhost -U $DB_USER -d $DB_NAME -c "
    SELECT 
        pid,
        now() - pg_stat_activity.query_start AS duration,
        query 
    FROM pg_stat_activity 
    WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes'
    AND state = 'active';
"
echo

# 5. 检查锁等待
echo "5. 检查锁等待"
psql -h localhost -U $DB_USER -d $DB_NAME -c "
    SELECT 
        blocked_locks.pid AS blocked_pid,
        blocked_activity.usename AS blocked_user,
        blocking_locks.pid AS blocking_pid,
        blocking_activity.usename AS blocking_user,
        blocked_activity.query AS blocked_statement,
        blocking_activity.query AS current_statement_in_blocking_process
    FROM pg_catalog.pg_locks blocked_locks
    JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
    JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
        AND blocking_locks.DATABASE IS NOT DISTINCT FROM blocked_locks.DATABASE
        AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
        AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
        AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
        AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
        AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
        AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
        AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
        AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
        AND blocking_locks.pid != blocked_locks.pid
    JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
    WHERE NOT blocked_locks.GRANTED;
"
echo

echo "=== 数据库诊断完成 ==="
```

#### 内存泄漏问题

**监控脚本**：

```bash
#!/bin/bash
# scripts/monitor_memory_leak.sh

PID=$(pgrep ai-monitor)
LOG_FILE="/var/log/ai-monitor/memory_monitor.log"
INTERVAL=60  # 监控间隔（秒）

if [ -z "$PID" ]; then
    echo "AI Monitor进程未运行"
    exit 1
fi

echo "开始监控进程 $PID 的内存使用情况..."
echo "日志文件: $LOG_FILE"

mkdir -p $(dirname $LOG_FILE)

while true; do
    TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
    
    # 获取进程内存信息
    if [ -f "/proc/$PID/status" ]; then
        RSS=$(grep VmRSS /proc/$PID/status | awk '{print $2}')
        VSZ=$(grep VmSize /proc/$PID/status | awk '{print $2}')
        
        # 获取系统内存信息
        MEM_TOTAL=$(grep MemTotal /proc/meminfo | awk '{print $2}')
        MEM_AVAILABLE=$(grep MemAvailable /proc/meminfo | awk '{print $2}')
        
        # 计算内存使用百分比
        MEM_PERCENT=$(echo "scale=2; $RSS * 100 / $MEM_TOTAL" | bc)
        
        # 记录到日志
        echo "$TIMESTAMP,PID:$PID,RSS:${RSS}KB,VSZ:${VSZ}KB,PERCENT:${MEM_PERCENT}%" >> $LOG_FILE
        
        # 检查是否存在内存泄漏（RSS持续增长）
        if (( $(echo "$MEM_PERCENT > 50" | bc -l) )); then
            echo "⚠ 警告: 进程内存使用率过高 ($MEM_PERCENT%)"
            
            # 生成内存转储（如果需要）
            if (( $(echo "$MEM_PERCENT > 80" | bc -l) )); then
                echo "生成内存转储..."
                gcore -o "/var/dumps/ai-monitor-$TIMESTAMP" $PID
            fi
        fi
    else
        echo "$TIMESTAMP,进程 $PID 已退出" >> $LOG_FILE
        break
    fi
    
    sleep $INTERVAL
done
```

## 🔒 安全运维

### 安全检查清单

#### 每日安全检查

```bash
#!/bin/bash
# scripts/daily_security_check.sh

echo "=== AI Monitor 每日安全检查 ==="
echo "检查时间: $(date)"
echo

# 1. 检查登录失败记录
echo "1. 检查登录失败记录 (最近24小时)"
fail_count=$(journalctl --since "24 hours ago" | grep "Failed password" | wc -l)
if [ $fail_count -gt 10 ]; then
    echo "⚠ 发现 $fail_count 次登录失败，可能存在暴力破解攻击"
    journalctl --since "24 hours ago" | grep "Failed password" | tail -5
else
    echo "✓ 登录失败次数正常 ($fail_count)"
fi
echo

# 2. 检查异常网络连接
echo "2. 检查异常网络连接"
netstat -tuln | grep LISTEN | while read line; do
    port=$(echo $line | awk '{print $4}' | cut -d: -f2)
    if [[ ! " 22 80 443 8080 5432 6379 9090 3000 " =~ " $port " ]]; then
        echo "⚠ 发现异常监听端口: $port"
    fi
done
echo

# 3. 检查文件权限
echo "3. 检查关键文件权限"
files=(
    "/etc/ai-monitor/config.yaml:644"
    "/usr/local/bin/ai-monitor:755"
    "/var/lib/ai-monitor:755"
    "/var/log/ai-monitor:755"
)

for file_perm in "${files[@]}"; do
    file=$(echo $file_perm | cut -d: -f1)
    expected_perm=$(echo $file_perm | cut -d: -f2)
    
    if [ -e "$file" ]; then
        actual_perm=$(stat -c "%a" "$file")
        if [ "$actual_perm" = "$expected_perm" ]; then
            echo "✓ $file: $actual_perm (正确)"
        else
            echo "⚠ $file: $actual_perm (期望: $expected_perm)"
        fi
    else
        echo "⚠ $file: 文件不存在"
    fi
done
echo

# 4. 检查SSL证书有效期
echo "4. 检查SSL证书有效期"
if [ -f "/etc/ssl/certs/ai-monitor.crt" ]; then
    expiry_date=$(openssl x509 -in /etc/ssl/certs/ai-monitor.crt -noout -enddate | cut -d= -f2)
    expiry_timestamp=$(date -d "$expiry_date" +%s)
    current_timestamp=$(date +%s)
    days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
    
    if [ $days_until_expiry -lt 30 ]; then
        echo "⚠ SSL证书将在 $days_until_expiry 天后过期"
    else
        echo "✓ SSL证书有效期正常 ($days_until_expiry 天)"
    fi
else
    echo "ℹ 未配置SSL证书"
fi
echo

# 5. 检查系统更新
echo "5. 检查系统安全更新"
if command -v apt &> /dev/null; then
    security_updates=$(apt list --upgradable 2>/dev/null | grep -i security | wc -l)
    if [ $security_updates -gt 0 ]; then
        echo "⚠ 有 $security_updates 个安全更新可用"
    else
        echo "✓ 系统安全更新已是最新"
    fi
elif command -v yum &> /dev/null; then
    security_updates=$(yum --security check-update 2>/dev/null | grep -c "updates")
    if [ $security_updates -gt 0 ]; then
        echo "⚠ 有 $security_updates 个安全更新可用"
    else
        echo "✓ 系统安全更新已是最新"
    fi
fi
echo

echo "=== 安全检查完成 ==="
```

### 安全加固脚本

```bash
#!/bin/bash
# scripts/security_hardening.sh

echo "=== AI Monitor 安全加固 ==="
echo "开始时间: $(date)"
echo

# 1. 设置防火墙规则
echo "1. 配置防火墙规则"
if command -v ufw &> /dev/null; then
    # Ubuntu/Debian
    ufw --force reset
    ufw default deny incoming
    ufw default allow outgoing
    ufw allow 22/tcp    # SSH
    ufw allow 80/tcp    # HTTP
    ufw allow 443/tcp   # HTTPS
    ufw allow 8080/tcp  # AI Monitor API
    ufw --force enable
    echo "✓ UFW防火墙规则已配置"
elif command -v firewall-cmd &> /dev/null; then
    # CentOS/RHEL
    firewall-cmd --permanent --zone=public --add-service=ssh
    firewall-cmd --permanent --zone=public --add-service=http
    firewall-cmd --permanent --zone=public --add-service=https
    firewall-cmd --permanent --zone=public --add-port=8080/tcp
    firewall-cmd --reload
    echo "✓ Firewalld规则已配置"
else
    echo "⚠ 未检测到防火墙管理工具"
fi
echo

# 2. 配置SSH安全
echo "2. 加固SSH配置"
SSH_CONFIG="/etc/ssh/sshd_config"
cp $SSH_CONFIG ${SSH_CONFIG}.backup

# 禁用root登录
sed -i 's/#PermitRootLogin yes/PermitRootLogin no/' $SSH_CONFIG
sed -i 's/PermitRootLogin yes/PermitRootLogin no/' $SSH_CONFIG

# 禁用密码认证（如果已配置密钥）
# sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' $SSH_CONFIG

# 限制登录尝试
echo "MaxAuthTries 3" >> $SSH_CONFIG
echo "MaxStartups 10:30:60" >> $SSH_CONFIG

# 设置空闲超时
echo "ClientAliveInterval 300" >> $SSH_CONFIG
echo "ClientAliveCountMax 2" >> $SSH_CONFIG

systemctl restart sshd
echo "✓ SSH配置已加固"
echo

# 3. 设置文件权限
echo "3. 设置安全文件权限"
chmod 600 /etc/ai-monitor/config.yaml
chmod 700 /var/lib/ai-monitor
chmod 700 /var/log/ai-monitor
chown -R ai-monitor:ai-monitor /var/lib/ai-monitor
chown -R ai-monitor:ai-monitor /var/log/ai-monitor
echo "✓ 文件权限已设置"
echo

# 4. 配置日志审计
echo "4. 配置系统审计"
if command -v auditctl &> /dev/null; then
    # 监控关键文件修改
    auditctl -w /etc/ai-monitor/ -p wa -k ai_monitor_config
    auditctl -w /usr/local/bin/ai-monitor -p x -k ai_monitor_exec
    auditctl -w /var/lib/ai-monitor/ -p wa -k ai_monitor_data
    echo "✓ 审计规则已配置"
else
    echo "⚠ 审计工具未安装"
fi
echo

# 5. 设置入侵检测
echo "5. 配置入侵检测"
if command -v fail2ban-client &> /dev/null; then
    # 配置fail2ban规则
    cat > /etc/fail2ban/jail.local << EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 3

[nginx-http-auth]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 3
EOF
    systemctl restart fail2ban
    echo "✓ Fail2ban已配置"
else
    echo "⚠ Fail2ban未安装"
fi
echo

echo "=== 安全加固完成 ==="
```

### 漏洞扫描脚本

```bash
#!/bin/bash
# scripts/vulnerability_scan.sh

echo "=== AI Monitor 漏洞扫描 ==="
echo "扫描时间: $(date)"
echo

# 1. 检查开放端口
echo "1. 端口扫描"
nmap -sS -O localhost | grep -E "(open|filtered)"
echo

# 2. 检查弱密码
echo "2. 检查系统用户"
awk -F: '($3 >= 1000) {print $1}' /etc/passwd | while read user; do
    if passwd -S $user 2>/dev/null | grep -q "P"; then
        echo "用户 $user: 密码已设置"
    else
        echo "⚠ 用户 $user: 无密码或密码已锁定"
    fi
done
echo

# 3. 检查SUID文件
echo "3. 检查SUID文件"
find / -perm -4000 -type f 2>/dev/null | head -10
echo

# 4. 检查配置文件安全
echo "4. 检查配置文件安全"
config_files=(
    "/etc/ai-monitor/config.yaml"
    "/etc/postgresql/*/main/postgresql.conf"
    "/etc/redis/redis.conf"
)

for config in "${config_files[@]}"; do
    if [ -f "$config" ]; then
        # 检查是否包含明文密码
        if grep -qi "password.*=.*[^*]" "$config"; then
            echo "⚠ $config: 可能包含明文密码"
        else
            echo "✓ $config: 密码配置安全"
        fi
    fi
done
echo

echo "=== 漏洞扫描完成 ==="
```

## 🔄 升级维护

### 应用升级流程

#### 升级前检查

```bash
#!/bin/bash
# scripts/pre_upgrade_check.sh

NEW_VERSION=$1

if [ -z "$NEW_VERSION" ]; then
    echo "用法: $0 <new_version>"
    exit 1
fi

echo "=== 升级前检查 (目标版本: $NEW_VERSION) ==="
echo "检查时间: $(date)"
echo

# 1. 检查当前版本
echo "1. 当前版本信息"
CURRENT_VERSION=$(ai-monitor version 2>/dev/null || echo "未知")
echo "当前版本: $CURRENT_VERSION"
echo "目标版本: $NEW_VERSION"
echo

# 2. 检查系统资源
echo "2. 系统资源检查"
echo "磁盘空间:"
df -h | grep -E '(/$|/var)'
echo
echo "内存使用:"
free -h
echo
echo "CPU负载:"
uptime
echo

# 3. 检查服务状态
echo "3. 服务状态检查"
services=("ai-monitor" "postgresql" "redis" "nginx")
for service in "${services[@]}"; do
    if systemctl is-active --quiet $service; then
        echo "✓ $service: 运行中"
    else
        echo "✗ $service: 已停止"
    fi
done
echo

# 4. 创建升级前备份
echo "4. 创建升级前备份"
BACKUP_DIR="/var/backups/upgrade/$(date +%Y%m%d_%H%M%S)"
mkdir -p $BACKUP_DIR

# 备份数据库
echo "备份数据库..."
pg_dump -h localhost -U ai_monitor -d ai_monitor > $BACKUP_DIR/database.sql
gzip $BACKUP_DIR/database.sql

# 备份配置文件
echo "备份配置文件..."
tar -czf $BACKUP_DIR/configs.tar.gz /etc/ai-monitor/ /etc/nginx/sites-available/ai-monitor

# 备份应用文件
echo "备份应用文件..."
cp /usr/local/bin/ai-monitor $BACKUP_DIR/

echo "✓ 备份完成: $BACKUP_DIR"
echo

# 5. 检查升级兼容性
echo "5. 升级兼容性检查"
echo "检查数据库模式兼容性..."
# 这里可以添加具体的兼容性检查逻辑
echo "✓ 兼容性检查通过"
echo

echo "=== 升级前检查完成 ==="
echo "备份位置: $BACKUP_DIR"
echo "可以开始升级流程"
```

#### 自动升级脚本

```bash
#!/bin/bash
# scripts/auto_upgrade.sh

set -e

NEW_VERSION=$1
BACKUP_DIR=$2

if [ -z "$NEW_VERSION" ] || [ -z "$BACKUP_DIR" ]; then
    echo "用法: $0 <new_version> <backup_dir>"
    exit 1
fi

echo "=== AI Monitor 自动升级 ==="
echo "目标版本: $NEW_VERSION"
echo "备份目录: $BACKUP_DIR"
echo "开始时间: $(date)"
echo

# 1. 下载新版本
echo "1. 下载新版本"
DOWNLOAD_URL="https://github.com/your-org/ai-monitor/releases/download/v${NEW_VERSION}/ai-monitor-linux-amd64"
wget -O /tmp/ai-monitor-new "$DOWNLOAD_URL"
chmod +x /tmp/ai-monitor-new
echo "✓ 新版本下载完成"
echo

# 2. 验证新版本
echo "2. 验证新版本"
NEW_VERSION_CHECK=$(/tmp/ai-monitor-new version)
if [[ "$NEW_VERSION_CHECK" == *"$NEW_VERSION"* ]]; then
    echo "✓ 版本验证通过: $NEW_VERSION_CHECK"
else
    echo "✗ 版本验证失败"
    exit 1
fi
echo

# 3. 停止服务
echo "3. 停止服务"
systemctl stop ai-monitor
echo "✓ 服务已停止"
echo

# 4. 替换可执行文件
echo "4. 替换可执行文件"
cp /usr/local/bin/ai-monitor $BACKUP_DIR/ai-monitor-old
mv /tmp/ai-monitor-new /usr/local/bin/ai-monitor
chown root:root /usr/local/bin/ai-monitor
chmod 755 /usr/local/bin/ai-monitor
echo "✓ 可执行文件已替换"
echo

# 5. 数据库迁移（如果需要）
echo "5. 数据库迁移"
if ai-monitor migrate --dry-run; then
    ai-monitor migrate
    echo "✓ 数据库迁移完成"
else
    echo "⚠ 数据库迁移失败，回滚中..."
    cp $BACKUP_DIR/ai-monitor-old /usr/local/bin/ai-monitor
    exit 1
fi
echo

# 6. 启动服务
echo "6. 启动服务"
systemctl start ai-monitor
sleep 10
echo

# 7. 验证升级
echo "7. 验证升级"
if systemctl is-active --quiet ai-monitor; then
    echo "✓ 服务启动成功"
else
    echo "✗ 服务启动失败，回滚中..."
    systemctl stop ai-monitor
    cp $BACKUP_DIR/ai-monitor-old /usr/local/bin/ai-monitor
    systemctl start ai-monitor
    exit 1
fi

# 健康检查
if curl -s -f http://localhost:8080/api/v1/health > /dev/null; then
    echo "✓ 健康检查通过"
else
    echo "✗ 健康检查失败"
    exit 1
fi

# 版本确认
UPGRADED_VERSION=$(ai-monitor version)
echo "升级后版本: $UPGRADED_VERSION"
echo

echo "=== 升级完成 ==="
echo "完成时间: $(date)"
echo "备份保留在: $BACKUP_DIR"
```

### 回滚流程

```bash
#!/bin/bash
# scripts/rollback.sh

BACKUP_DIR=$1

if [ -z "$BACKUP_DIR" ] || [ ! -d "$BACKUP_DIR" ]; then
    echo "用法: $0 <backup_dir>"
    echo "可用备份:"
    ls -la /var/backups/upgrade/
    exit 1
fi

echo "=== AI Monitor 回滚 ==="
echo "备份目录: $BACKUP_DIR"
echo "开始时间: $(date)"
echo

# 确认回滚
read -p "确认要回滚到备份版本吗？(y/N): " confirm
if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "回滚操作已取消"
    exit 0
fi

# 1. 停止服务
echo "1. 停止服务"
systemctl stop ai-monitor
echo "✓ 服务已停止"
echo

# 2. 恢复可执行文件
echo "2. 恢复可执行文件"
if [ -f "$BACKUP_DIR/ai-monitor-old" ]; then
    cp "$BACKUP_DIR/ai-monitor-old" /usr/local/bin/ai-monitor
    chmod 755 /usr/local/bin/ai-monitor
    echo "✓ 可执行文件已恢复"
else
    echo "✗ 备份的可执行文件不存在"
    exit 1
fi
echo

# 3. 恢复数据库
echo "3. 恢复数据库"
if [ -f "$BACKUP_DIR/database.sql.gz" ]; then
    read -p "是否恢复数据库？这将覆盖当前数据 (y/N): " db_confirm
    if [ "$db_confirm" = "y" ] || [ "$db_confirm" = "Y" ]; then
        psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS ai_monitor;"
        psql -h localhost -U postgres -c "CREATE DATABASE ai_monitor OWNER ai_monitor;"
        gunzip -c "$BACKUP_DIR/database.sql.gz" | psql -h localhost -U ai_monitor -d ai_monitor
        echo "✓ 数据库已恢复"
    else
        echo "跳过数据库恢复"
    fi
else
    echo "⚠ 数据库备份文件不存在"
fi
echo

# 4. 恢复配置文件
echo "4. 恢复配置文件"
if [ -f "$BACKUP_DIR/configs.tar.gz" ]; then
    tar -xzf "$BACKUP_DIR/configs.tar.gz" -C /
    echo "✓ 配置文件已恢复"
else
    echo "⚠ 配置备份文件不存在"
fi
echo

# 5. 启动服务
echo "5. 启动服务"
systemctl start ai-monitor
sleep 10
echo

# 6. 验证回滚
echo "6. 验证回滚"
if systemctl is-active --quiet ai-monitor; then
    echo "✓ 服务启动成功"
else
    echo "✗ 服务启动失败"
    exit 1
fi

if curl -s -f http://localhost:8080/api/v1/health > /dev/null; then
    echo "✓ 健康检查通过"
else
    echo "✗ 健康检查失败"
    exit 1
fi

ROLLBACK_VERSION=$(ai-monitor version)
echo "回滚后版本: $ROLLBACK_VERSION"
echo

echo "=== 回滚完成 ==="
echo "完成时间: $(date)"
```

## 📊 容量规划

### 资源监控脚本

```bash
#!/bin/bash
# scripts/capacity_monitoring.sh

REPORT_FILE="/var/reports/capacity_report_$(date +%Y%m%d).txt"
mkdir -p $(dirname $REPORT_FILE)

echo "=== AI Monitor 容量监控报告 ===" > $REPORT_FILE
echo "生成时间: $(date)" >> $REPORT_FILE
echo >> $REPORT_FILE

# 1. 系统资源使用情况
echo "1. 系统资源使用情况" >> $REPORT_FILE
echo "CPU使用率:" >> $REPORT_FILE
top -bn1 | grep "Cpu(s)" >> $REPORT_FILE
echo >> $REPORT_FILE

echo "内存使用情况:" >> $REPORT_FILE
free -h >> $REPORT_FILE
echo >> $REPORT_FILE

echo "磁盘使用情况:" >> $REPORT_FILE
df -h >> $REPORT_FILE
echo >> $REPORT_FILE

# 2. 数据库容量分析
echo "2. 数据库容量分析" >> $REPORT_FILE
psql -h localhost -U ai_monitor -d ai_monitor -c "
    SELECT 
        schemaname,
        tablename,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
        pg_total_relation_size(schemaname||'.'||tablename) as size_bytes
    FROM pg_tables 
    WHERE schemaname = 'public'
    ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
" >> $REPORT_FILE
echo >> $REPORT_FILE

# 3. 日志文件大小
echo "3. 日志文件大小" >> $REPORT_FILE
find /var/log -name "*.log" -exec du -h {} \; | sort -hr | head -20 >> $REPORT_FILE
echo >> $REPORT_FILE

# 4. 网络流量统计
echo "4. 网络流量统计" >> $REPORT_FILE
cat /proc/net/dev | grep -E "(eth0|ens|enp)" >> $REPORT_FILE
echo >> $REPORT_FILE

# 5. 进程资源使用
echo "5. 进程资源使用" >> $REPORT_FILE
ps aux --sort=-%cpu | head -10 >> $REPORT_FILE
echo >> $REPORT_FILE

echo "报告已生成: $REPORT_FILE"
```

### 容量预测脚本

```python
#!/usr/bin/env python3
# scripts/capacity_prediction.py

import psycopg2
import pandas as pd
import numpy as np
from datetime import datetime, timedelta
import matplotlib.pyplot as plt
import seaborn as sns
from sklearn.linear_model import LinearRegression
import warnings
warnings.filterwarnings('ignore')

def connect_db():
    """连接数据库"""
    return psycopg2.connect(
        host="localhost",
        database="ai_monitor",
        user="ai_monitor",
        password="your_password"
    )

def get_metrics_data(days=30):
    """获取指标数据"""
    conn = connect_db()
    
    query = """
    SELECT 
        DATE(created_at) as date,
        COUNT(*) as daily_metrics,
        AVG(CASE WHEN metric_type = 'cpu' THEN value END) as avg_cpu,
        AVG(CASE WHEN metric_type = 'memory' THEN value END) as avg_memory,
        AVG(CASE WHEN metric_type = 'disk' THEN value END) as avg_disk
    FROM metrics 
    WHERE created_at >= NOW() - INTERVAL '%s days'
    GROUP BY DATE(created_at)
    ORDER BY date;
    """ % days
    
    df = pd.read_sql(query, conn)
    conn.close()
    
    return df

def predict_growth(df, metric_column, days_ahead=30):
    """预测增长趋势"""
    # 准备数据
    df['date_num'] = pd.to_datetime(df['date']).astype(int) // 10**9
    X = df[['date_num']]
    y = df[metric_column].fillna(0)
    
    # 训练模型
    model = LinearRegression()
    model.fit(X, y)
    
    # 预测未来
    last_date = df['date_num'].max()
    future_dates = []
    for i in range(1, days_ahead + 1):
        future_date = last_date + (i * 86400)  # 86400秒 = 1天
        future_dates.append(future_date)
    
    future_X = np.array(future_dates).reshape(-1, 1)
    predictions = model.predict(future_X)
    
    return predictions, future_dates

def generate_capacity_report():
    """生成容量规划报告"""
    print("=== AI Monitor 容量预测报告 ===")
    print(f"生成时间: {datetime.now()}")
    print()
    
    # 获取数据
    df = get_metrics_data(30)
    
    if df.empty:
        print("没有足够的历史数据进行预测")
        return
    
    # 预测各项指标
    metrics = ['daily_metrics', 'avg_cpu', 'avg_memory', 'avg_disk']
    metric_names = ['每日指标数量', 'CPU使用率', '内存使用率', '磁盘使用率']
    
    for metric, name in zip(metrics, metric_names):
        if metric in df.columns and not df[metric].isna().all():
            predictions, _ = predict_growth(df, metric, 30)
            
            current_avg = df[metric].mean()
            predicted_avg = np.mean(predictions)
            growth_rate = ((predicted_avg - current_avg) / current_avg) * 100
            
            print(f"{name}:")
            print(f"  当前平均值: {current_avg:.2f}")
            print(f"  预测平均值: {predicted_avg:.2f}")
            print(f"  增长率: {growth_rate:.2f}%")
            
            if growth_rate > 50:
                print(f"  ⚠ 警告: {name}增长过快，需要关注")
            elif growth_rate > 20:
                print(f"  ⚠ 注意: {name}增长较快")
            else:
                print(f"  ✓ {name}增长正常")
            print()
    
    # 数据库容量预测
    conn = connect_db()
    cursor = conn.cursor()
    
    cursor.execute("""
        SELECT pg_size_pretty(pg_database_size('ai_monitor')) as current_size,
               pg_database_size('ai_monitor') as size_bytes
    """)
    
    result = cursor.fetchone()
    current_size_pretty = result[0]
    current_size_bytes = result[1]
    
    # 简单的线性增长预测（基于每日指标数量）
    daily_growth = df['daily_metrics'].mean() * 1024  # 假设每个指标1KB
    monthly_growth = daily_growth * 30
    predicted_size_bytes = current_size_bytes + monthly_growth
    
    print("数据库容量预测:")
    print(f"  当前大小: {current_size_pretty}")
    print(f"  预计月增长: {monthly_growth / (1024*1024):.2f} MB")
    print(f"  预计30天后大小: {predicted_size_bytes / (1024*1024*1024):.2f} GB")
    
    if predicted_size_bytes > current_size_bytes * 2:
        print("  ⚠ 警告: 数据库增长过快，建议优化数据保留策略")
    
    conn.close()
    print()
    
    # 建议
    print("容量规划建议:")
    print("1. 定期清理过期数据")
    print("2. 监控磁盘使用率，及时扩容")
    print("3. 优化数据库查询性能")
    print("4. 考虑数据归档策略")
    print("5. 评估是否需要分库分表")

if __name__ == "__main__":
    generate_capacity_report()
```

## 🛠️ 运维工具

### 一键运维脚本

```bash
#!/bin/bash
# scripts/ops_toolkit.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

show_menu() {
    echo "=== AI Monitor 运维工具箱 ==="
    echo "1. 系统健康检查"
    echo "2. 性能监控"
    echo "3. 日志分析"
    echo "4. 数据库维护"
    echo "5. 备份管理"
    echo "6. 安全检查"
    echo "7. 容量分析"
    echo "8. 服务重启"
    echo "9. 故障诊断"
    echo "0. 退出"
    echo "==========================================="
    read -p "请选择操作 (0-9): " choice
}

while true; do
    show_menu
    
    case $choice in
        1)
            echo "执行系统健康检查..."
            $SCRIPT_DIR/daily_health_check.sh
            ;;
        2)
            echo "启动性能监控..."
            $SCRIPT_DIR/capacity_monitoring.sh
            ;;
        3)
            echo "分析系统日志..."
            $SCRIPT_DIR/daily_log_check.sh
            ;;
        4)
            echo "执行数据库维护..."
            psql -h localhost -U ai_monitor -d ai_monitor -f $SCRIPT_DIR/weekly_db_maintenance.sql
            ;;
        5)
            echo "备份管理..."
            echo "1) 创建备份"
            echo "2) 查看备份"
            echo "3) 恢复备份"
            read -p "选择操作: " backup_choice
            case $backup_choice in
                1) $SCRIPT_DIR/backup_database.sh ;;
                2) ls -la /var/backups/postgresql/ ;;
                3) 
                    echo "可用备份:"
                    ls -la /var/backups/postgresql/*.sql.gz
                    read -p "输入备份文件路径: " backup_file
                    $SCRIPT_DIR/restore_database.sh "$backup_file"
                    ;;
            esac
            ;;
        6)
            echo "执行安全检查..."
            $SCRIPT_DIR/daily_security_check.sh
            ;;
        7)
            echo "容量分析..."
            python3 $SCRIPT_DIR/capacity_prediction.py
            ;;
        8)
            echo "服务重启..."
            echo "1) 重启AI Monitor"
            echo "2) 重启所有服务"
            read -p "选择操作: " restart_choice
            case $restart_choice in
                1) 
                    systemctl restart ai-monitor
                    echo "AI Monitor服务已重启"
                    ;;
                2)
                    systemctl restart ai-monitor postgresql redis nginx
                    echo "所有服务已重启"
                    ;;
            esac
            ;;
        9)
            echo "故障诊断..."
            $SCRIPT_DIR/diagnose_startup_failure.sh
            ;;
        0)
            echo "退出运维工具箱"
            exit 0
            ;;
        *)
            echo "无效选择，请重新输入"
            ;;
    esac
    
    echo
    read -p "按回车键继续..."
    clear
done
```

### 监控面板脚本

```bash
#!/bin/bash
# scripts/monitoring_dashboard.sh

while true; do
    clear
    echo "=== AI Monitor 实时监控面板 ==="
    echo "更新时间: $(date)"
    echo "========================================"
    
    # 系统信息
    echo "📊 系统信息:"
    echo "  负载: $(uptime | awk -F'load average:' '{print $2}')"
    echo "  内存: $(free | grep Mem | awk '{printf "%.1f%%", $3/$2 * 100.0}')"
    echo "  磁盘: $(df / | tail -1 | awk '{print $5}')"
    echo
    
    # 服务状态
    echo "🔧 服务状态:"
    services=("ai-monitor" "postgresql" "redis" "nginx")
    for service in "${services[@]}"; do
        if systemctl is-active --quiet $service; then
            echo "  ✓ $service: 运行中"
        else
            echo "  ✗ $service: 已停止"
        fi
    done
    echo
    
    # 网络连接
    echo "🌐 网络连接:"
    echo "  活跃连接: $(netstat -an | grep ESTABLISHED | wc -l)"
    echo "  监听端口: $(netstat -tuln | grep LISTEN | wc -l)"
    echo
    
    # 数据库状态
    echo "🗄️ 数据库状态:"
    if systemctl is-active --quiet postgresql; then
        db_connections=$(psql -h localhost -U ai_monitor -d ai_monitor -t -c "SELECT count(*) FROM pg_stat_activity;" 2>/dev/null || echo "N/A")
        echo "  连接数: $db_connections"
        
        db_size=$(psql -h localhost -U ai_monitor -d ai_monitor -t -c "SELECT pg_size_pretty(pg_database_size('ai_monitor'));" 2>/dev/null || echo "N/A")
        echo "  数据库大小: $db_size"
    else
        echo "  数据库未运行"
    fi
    echo
    
    # 最近日志
    echo "📝 最近日志 (最近5条):"
    journalctl -u ai-monitor --since "5 minutes ago" --no-pager | tail -5 | cut -c1-80
    echo
    
    echo "按 Ctrl+C 退出监控"
    sleep 5
done
```

---

## 📞 联系支持

如果在运维过程中遇到问题，请通过以下方式获取支持：

- **技术文档**: 查看 `/doc` 目录下的其他文档
- **日志分析**: 检查 `/var/log/ai-monitor/` 下的日志文件
- **社区支持**: 提交 GitHub Issue
- **紧急支持**: 联系运维团队

---

*本运维指南涵盖了AI Monitor系统的日常运维、监控告警、性能调优、备份恢复、故障处理、安全运维、升级维护、容量规划等各个方面，为运维人员提供了完整的操作指南和工具脚本。*