# AI Monitor 镜像构建脚本使用说明

## 问题解决

如果您遇到了以下错误：
```
ERROR: failed to build: failed to solve: dockerfile parse error on line 23: unknown instruction: for (did you mean from?)
```

这通常是因为以下原因：

### 1. 执行目录错误

**错误做法：**
```powershell
# 在其他目录下执行脚本
C:\AItest> .\build_images.ps1
```

**正确做法：**
```powershell
# 必须在项目根目录下执行
cd C:\Users\Administrator\Downloads\GoCode\AIMonitor
.\build_images.ps1
```

### 2. 脚本必须在项目根目录执行

构建脚本需要访问以下文件和目录：
- `go.mod` 和 `go.sum`
- `cmd/server/main.go`
- `web/` 目录
- `configs/` 目录

这些文件只存在于项目根目录中。

## 正确的使用步骤

### 步骤 1：切换到项目根目录
```powershell
cd C:\Users\Administrator\Downloads\GoCode\AIMonitor
```

### 步骤 2：确认项目结构
确保当前目录包含以下文件：
- `go.mod`
- `cmd/server/main.go`
- `web/package.json`
- `build_images.ps1`

### 步骤 3：执行构建脚本
```powershell
powershell -ExecutionPolicy Bypass -File .\build_images.ps1
```

或者直接：
```powershell
.\build_images.ps1
```

## 构建过程说明

脚本将执行以下步骤：
1. 检查项目结构
2. 检查 Docker 环境
3. 创建输出目录 `./docker-images`
4. 清理 Docker 缓存
5. 创建优化的 Dockerfile
6. 构建后端镜像
7. 构建前端镜像
8. 保存镜像到文件
9. 创建镜像清单和加载脚本

## 输出结果

构建成功后，将在 `./docker-images` 目录下生成：
- `ai-monitor-backend.tar.gz` - 后端服务镜像
- `ai-monitor-frontend.tar.gz` - 前端服务镜像
- `postgres.tar.gz` - PostgreSQL 数据库镜像
- `redis.tar.gz` - Redis 缓存镜像
- `prometheus.tar.gz` - Prometheus 监控镜像
- `elasticsearch.tar.gz` - Elasticsearch 搜索引擎镜像
- `load_images.sh` - 镜像加载脚本
- `images.txt` - 镜像清单文件

## 常见问题

### Q: 为什么不能在其他目录执行？
A: 脚本需要读取项目的 `go.mod`、源代码文件等，这些文件的路径是相对于项目根目录的。

### Q: 如何确认当前在正确的目录？
A: 执行 `ls` 或 `dir` 命令，应该能看到 `go.mod`、`cmd`、`web` 等文件和目录。

### Q: 构建失败怎么办？
A: 检查 Docker 是否正常运行，网络是否正常，以及是否在正确的目录下执行脚本。

## 下一步

构建完成后，可以：
1. 将 `docker-images` 目录传输到远程服务器
2. 在远程服务器上运行 `./load_images.sh` 加载镜像
3. 运行 `deploy_from_images.sh` 部署服务

或者使用一键部署脚本：
```powershell
.\quick_deploy.ps1
```