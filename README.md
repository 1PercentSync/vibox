# ViBox

> 基于 Docker 的 Web 工作空间管理系统

## 项目简介

ViBox 是一个通过 Web 界面管理 Docker 容器工作空间的系统，让用户能够：
- 通过浏览器创建和管理隔离的开发环境
- 在浏览器中访问容器终端（WebSSH）
- 访问容器内运行的 HTTP 服务（端口转发）
- 使用自定义脚本初始化工作空间

## 项目状态

🔄 **当前阶段**：第一阶段 - Go 后端开发

| 阶段 | 状态 | 说明 |
|------|------|------|
| 第一阶段 | 🔄 进行中 | Go 后端核心功能 |
| 第二阶段 | ⏳ 待定 | 前端界面 + MVP 集成 |
| 第三阶段 | ⏳ 待定 | 完整功能扩展 |

## 快速开始

> 注意：项目正在开发中，以下内容为计划中的使用方式

### 部署

```bash
# 1. 设置 API Token（必须）
export API_TOKEN=$(openssl rand -hex 32)

# 2. 配置 docker-compose.yml，添加环境变量：
# environment:
#   - API_TOKEN=your-secret-token

# 3. 启动服务
docker-compose up -d

# 4. 访问（需要 token）
# http://localhost:3000
```

### 配置 Caddy 反向代理

```
# Caddyfile
your-domain.com {
    reverse_proxy localhost:3000
}
```

## 核心功能

### 第一阶段（当前开发）

- ✅ **Token 鉴权**（环境变量配置）
- ✅ Docker 容器管理（创建、启动、停止、删除、重置）
- ✅ 自定义脚本执行
- ✅ WebSSH 终端访问
- ✅ HTTP 端口转发（动态访问 + 端口标签）
- ✅ 数据持久化（工作空间配置自动恢复）

### 第二阶段（计划）

- ⏳ Web 前端界面
- ⏳ 工作空间可视化管理
- ⏳ 脚本管理界面
- ⏳ 端口转发控制面板

### 第三阶段（计划）

- ⏳ GitHub 集成
- ⏳ AI Coding Agent 集成
- ⏳ VS Code Server 集成
- ⏳ 用户认证与权限管理

## 技术栈

### 后端（第一阶段）

- **语言**：Go 1.25+
- **Web 框架**：Gin
- **Docker SDK**：github.com/docker/docker/client
- **WebSocket**：github.com/gorilla/websocket
- **反向代理**：net/http/httputil（标准库）

### 前端（待定）

- React/Vue + TypeScript + xterm.js（待定）

## 文档

- [项目路线图](./PROJECT_ROADMAP.md) - 三个阶段的详细规划
- [第一阶段：后端实现](./docs/PHASE1_BACKEND.md) - 当前阶段的详细技术文档
- [任务拆分方案](./docs/PHASE1_TASK_BREAKDOWN.md) - 模块化开发和并行任务分配
- [API 规范](./docs/API_SPECIFICATION.md) - RESTful API 和 WebSocket 接口定义

## 架构

### 第一阶段架构

```
用户浏览器
    ↓
Caddy (domain.com)
    ↓
反向代理到 localhost:3000
    ↓
┌─────────────────────────────────────┐
│  Go 后端服务 (端口 3000)             │
│  ├── /api/*      RESTful API        │
│  ├── /ws/*       WebSocket 终端     │
│  └── /forward/*  端口转发           │
└──────────────┬──────────────────────┘
               │
        Docker Engine
        └── 工作空间容器
```

## API 示例

> **注意**：所有 API 都需要 Token 鉴权

### 创建工作空间

```bash
# 使用 Authorization Header（推荐）
curl -X POST http://localhost:3000/api/workspaces \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret-token" \
  -d '{
    "name": "my-workspace",
    "image": "ubuntu:22.04",
    "scripts": [
      {
        "name": "install-tools",
        "content": "#!/bin/bash\napt-get update && apt-get install -y curl git",
        "order": 1
      }
    ],
    "ports": {
      "8080": "VS Code Server",
      "3000": "Web App"
    }
  }'

# 或使用查询参数
curl -X POST "http://localhost:3000/api/workspaces?token=your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```

### 访问终端

```javascript
// WebSocket 连接需要在 URL 中携带 token
const ws = new WebSocket('ws://localhost:3000/ws/terminal/{workspace-id}?token=your-secret-token');
ws.onmessage = (event) => console.log(event.data);
ws.send(JSON.stringify({type: 'input', data: 'ls -la\n'}));
```

### 访问容器内 HTTP 服务

```bash
# 容器内运行的服务在 8080 端口
# 通过以下 URL 访问（需要 token）：
http://localhost:3000/forward/{workspace-id}/8080/?token=your-secret-token
```

## 开发

### 环境要求

- Go 1.25+
- Docker
- Git

### 本地开发

```bash
# 克隆项目
git clone https://github.com/1PercentSync/vibox.git
cd vibox

# 安装依赖
go mod download

# 设置 API Token（必须）
export API_TOKEN=dev-token-123

# 运行
go run ./cmd/server
```

### 开发进度

参见 [第一阶段开发计划](./docs/PHASE1_BACKEND.md#开发计划)

## 贡献

项目正在开发中，欢迎贡献！

## 许可证

待定

## 联系方式

- GitHub: https://github.com/1PercentSync/vibox
- Issue: https://github.com/1PercentSync/vibox/issues
