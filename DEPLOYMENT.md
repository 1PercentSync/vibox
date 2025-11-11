# ViBox 部署指南

> **部署方式**：ViBox 仅支持 Docker 部署

---

## 目录

1. [快速开始](#快速开始)
2. [使用 Docker Compose 部署（推荐）](#使用-docker-compose-部署推荐)
3. [使用 Docker 手动部署](#使用-docker-手动部署)
4. [从源码构建镜像](#从源码构建镜像)
5. [环境变量配置](#环境变量配置)
6. [生产环境部署](#生产环境部署)
7. [故障排除](#故障排除)

---

## 快速开始

### 前置条件

- Docker 20.10+
- Docker Compose v2.0+（推荐）
- Git

### 快速部署

```bash
# 1. 克隆仓库
git clone https://github.com/1PercentSync/vibox.git
cd vibox

# 2. 创建环境变量文件
cp .env.example .env

# 3. 生成 API Token
echo "API_TOKEN=$(openssl rand -hex 32)" > .env

# 4. 启动服务
docker-compose up -d

# 5. 检查服务状态
docker-compose ps
curl http://localhost:3000/health

# 6. 访问应用
# 浏览器打开: http://localhost:3000
```

---

## 使用 Docker Compose 部署（推荐）

### 1. 准备配置

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件
nano .env
```

**最小配置示例**：
```env
API_TOKEN=your-secret-token-here
PORT=3000
```

### 2. 启动服务

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f vibox

# 停止服务
docker-compose down

# 停止并删除数据
docker-compose down -v
```

### 3. 验证部署

```bash
# 健康检查
curl http://localhost:3000/health

# 登录并测试 API
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"token": "your-token"}'

# 访问 API（Cookie 自动发送）
curl -b cookies.txt http://localhost:3000/api/workspaces
```

---

## 使用 Docker 手动部署

### 1. 拉取镜像

```bash
# 从 GitHub Container Registry 拉取
docker pull ghcr.io/1percentsync/vibox:latest

# 或指定版本
docker pull ghcr.io/1percentsync/vibox:v1.0.0
```

### 2. 运行容器

```bash
docker run -d \
  --name vibox \
  --restart unless-stopped \
  -p 3000:3000 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -e API_TOKEN="$(openssl rand -hex 32)" \
  -e PORT=3000 \
  -e DEFAULT_IMAGE=ubuntu:22.04 \
  ghcr.io/1percentsync/vibox:latest
```

### 3. 管理容器

```bash
# 查看日志
docker logs -f vibox

# 停止容器
docker stop vibox

# 启动容器
docker start vibox

# 删除容器
docker rm -f vibox
```

---

## 从源码构建镜像

### 构建说明

ViBox 使用**多阶段 Docker 构建**（Phase 2.5），自动完成：
1. **Stage 1**: 构建 React 前端（Node.js）
2. **Stage 2**: 构建 Go 后端并嵌入前端
3. **Stage 3**: 创建最小运行时镜像

### 1. 构建本地镜像

```bash
# 克隆仓库
git clone https://github.com/1PercentSync/vibox.git
cd vibox

# 构建镜像
docker build -t vibox:local .

# 查看镜像
docker images | grep vibox
```

### 2. 多平台构建

```bash
# 使用 buildx 构建多平台镜像
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t vibox:local \
  .
```

### 3. 使用本地镜像运行

```bash
# 使用本地构建的镜像
docker run -d \
  --name vibox \
  -p 3000:3000 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -e API_TOKEN="your-secret-token" \
  vibox:local
```

### 构建架构

```
┌────────────────────────────────────────┐
│ Stage 1: Frontend Builder (Node.js)   │
│ - npm ci                               │
│ - npm run build                        │
│ → frontend/dist/                       │
└────────────────┬───────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────┐
│ Stage 2: Backend Builder (Go)         │
│ - go mod download                      │
│ - Copy frontend/dist → static/dist     │
│ - go build (embeds frontend)           │
│ → /build/vibox                         │
└────────────────┬───────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────┐
│ Stage 3: Runtime (Alpine)             │
│ - Copy vibox binary                    │
│ - Minimal runtime dependencies         │
│ → Final image (~30-40MB)               │
└────────────────────────────────────────┘
```

### 构建优化

Docker 构建过程已优化：
- ✅ 利用 Docker 层缓存
- ✅ 前端构建在独立阶段
- ✅ 最小运行时镜像（Alpine Linux）
- ✅ 静态二进制文件（无 CGO 依赖）
- ✅ 去除调试符号（-ldflags="-s -w"）

**镜像大小**：
- Frontend Builder: ~500MB（仅构建阶段）
- Backend Builder: ~1GB（仅构建阶段）
- **最终运行时镜像: ~30-40MB** ✅

---

## 环境变量配置

### 必需配置

| 变量 | 说明 | 示例 |
|------|------|------|
| `API_TOKEN` | API 鉴权令牌 | `openssl rand -hex 32` |

### 可选配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PORT` | 服务监听端口 | `3000` |
| `DOCKER_HOST` | Docker 守护进程地址 | `unix:///var/run/docker.sock` |
| `DEFAULT_IMAGE` | 默认容器镜像 | `ubuntu:22.04` |
| `MEMORY_LIMIT` | 容器内存限制（字节） | `536870912` (512MB) |
| `CPU_LIMIT` | 容器 CPU 限制（纳秒） | `1000000000` (1 CPU) |

### 生成安全的 API Token

```bash
# 方法 1: OpenSSL
openssl rand -hex 32

# 方法 2: UUID
uuidgen

# 方法 3: /dev/urandom
head -c 32 /dev/urandom | base64
```

---

## 生产环境部署

### 1. 使用反向代理（推荐）

#### Caddy 配置

```caddyfile
# Caddyfile
your-domain.com {
    reverse_proxy localhost:3000

    # 可选: 速率限制
    rate_limit {
        zone dynamic {
            key {remote_host}
            events 100
            window 1m
        }
    }
}
```

启动 Caddy：
```bash
caddy run --config Caddyfile
```

#### Nginx 配置

```nginx
# /etc/nginx/sites-available/vibox
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2. 安全配置

**Docker Socket 权限**：
```bash
# 创建 docker 组用户（推荐）
sudo groupadd docker
sudo usermod -aG docker vibox-user

# 或使用只读挂载（更安全，但功能受限）
-v /var/run/docker.sock:/var/run/docker.sock:ro
```

**防火墙配置**：
```bash
# 仅允许通过反向代理访问
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw deny 3000/tcp  # 阻止直接访问
```

### 3. 监控和日志

**健康检查**：
```bash
# 使用 cron 定期检查
*/5 * * * * curl -f http://localhost:3000/health || systemctl restart docker-vibox
```

**日志轮转**：
```yaml
# docker-compose.yml 中配置
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

### 4. 备份和恢复

ViBox 当前使用内存存储，重启后数据会丢失。如需持久化：

1. **工作空间列表**：定期导出
2. **容器数据**：使用 Docker volumes
3. **配置文件**：备份 `.env` 文件

---

## 故障排除

### 服务无法启动

**问题**: `API_TOKEN is required`

**解决**:
```bash
# 确保设置了 API_TOKEN
echo "API_TOKEN=$(openssl rand -hex 32)" >> .env
docker-compose up -d
```

### Docker Socket 权限错误

**问题**: `permission denied while trying to connect to the Docker daemon socket`

**解决**:
```bash
# 方法 1: 添加用户到 docker 组
sudo usermod -aG docker $USER
newgrp docker

# 方法 2: 修改 socket 权限（不推荐生产环境）
sudo chmod 666 /var/run/docker.sock
```

### 端口已被占用

**问题**: `bind: address already in use`

**解决**:
```bash
# 查找占用端口的进程
sudo lsof -i :3000

# 修改 .env 中的 PORT 或 HOST_PORT
echo "HOST_PORT=3001" >> .env
docker-compose up -d
```

### 健康检查失败

**问题**: `curl: (7) Failed to connect to localhost port 3000`

**解决**:
```bash
# 1. 查看容器日志
docker-compose logs vibox

# 2. 检查容器状态
docker-compose ps

# 3. 验证配置
docker exec vibox env | grep API_TOKEN

# 4. 重启服务
docker-compose restart
```

### WebSocket 连接失败

**问题**: 终端无法连接

**解决**:
1. 确保反向代理支持 WebSocket 升级
2. 检查防火墙规则
3. 验证 Token 鉴权正确

**Nginx WebSocket 配置**:
```nginx
proxy_http_version 1.1;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
```

---

## 更新和升级

### 使用 Docker Compose

```bash
# 拉取最新镜像
docker-compose pull

# 重启服务
docker-compose up -d

# 清理旧镜像
docker image prune -f
```

### 使用 Docker 手动更新

```bash
# 停止并删除旧容器
docker stop vibox
docker rm vibox

# 拉取新镜像
docker pull ghcr.io/1percentsync/vibox:latest

# 启动新容器（使用相同配置）
docker run -d \
  --name vibox \
  --restart unless-stopped \
  -p 3000:3000 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -e API_TOKEN="your-token" \
  ghcr.io/1percentsync/vibox:latest
```

---

## 性能优化

### 1. 资源限制

```yaml
# docker-compose.yml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
    reservations:
      cpus: '1'
      memory: 512M
```

### 2. Docker 优化

```bash
# 清理未使用的容器
docker system prune -af

# 限制日志大小
docker-compose.yml:
  logging:
    options:
      max-size: "10m"
      max-file: "3"
```

---

## 支持

- **GitHub Issues**: https://github.com/1PercentSync/vibox/issues
- **文档**: https://github.com/1PercentSync/vibox/blob/main/README.md

---

**部署成功后，访问**：
- 应用: http://localhost:3000
- 健康检查: http://localhost:3000/health
- API 文档: 参见 [API_SPECIFICATION.md](./docs/API_SPECIFICATION.md)
