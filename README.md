# ViBox

> 基于 Docker 的 Web 工作空间管理系统

## 功能

- 通过浏览器创建和管理 Docker 容器工作空间
- Web 终端访问（WebSocket）
- HTTP 端口转发
- 自定义初始化脚本

## 快速开始

### 部署

```bash
# 克隆仓库
git clone https://github.com/1PercentSync/vibox.git
cd vibox

# 设置环境变量
echo "API_TOKEN=$(openssl rand -hex 32)" > .env

# 启动
docker-compose up -d

# 访问 http://localhost:3000
```

详细部署说明：[DEPLOYMENT.md](./DEPLOYMENT.md)

## 技术栈

**后端**: Go + Gin + Docker SDK + WebSocket
**前端**: React + TypeScript + Vite + Tailwind CSS + xterm.js

## 架构

```
浏览器
  ↓
Caddy/Nginx (可选)
  ↓
ViBox Docker 容器 (:3000)
  ├── /api/*      RESTful API
  ├── /ws/*       WebSocket 终端
  ├── /forward/*  端口转发
  └── /           React 前端
  ↓
Docker Engine
  └── 工作空间容器
```

## API 示例

```bash
# 登录
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"token": "your-secret-token"}'

# 创建工作空间
curl -X POST http://localhost:3000/api/workspaces \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-workspace",
    "image": "ubuntu:22.04",
    "ports": {
      "8080": "VS Code Server"
    }
  }'
```

完整 API 文档：[docs/API_SPECIFICATION.md](./docs/API_SPECIFICATION.md)

## 开发

### 环境要求

- Docker 20.10+
- Docker Compose v2.0+

### 本地构建

```bash
# 构建 Docker 镜像
docker build -t vibox:local .

# 运行
docker run -d \
  --name vibox \
  -p 3000:3000 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -e API_TOKEN="your-secret-token" \
  vibox:local
```

### 前端开发（可选）

如需修改前端代码：

```bash
cd frontend
npm install
npm run dev  # 开发服务器运行在 :5173
```

Vite 会自动代理 API 请求到后端 `:3000`。

## 许可证

MIT

## 联系方式

- GitHub: https://github.com/1PercentSync/vibox
- Issues: https://github.com/1PercentSync/vibox/issues
