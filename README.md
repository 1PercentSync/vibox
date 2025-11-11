# ViBox

> 基于 Docker 的 Web 工作空间管理系统

## 项目简介

ViBox 是一个通过 Web 界面管理 Docker 容器工作空间的系统，让用户能够：
- 通过浏览器创建和管理隔离的开发环境
- 在浏览器中访问容器终端（WebSSH）
- 访问容器内运行的 HTTP 服务（端口转发）
- 使用自定义脚本初始化工作空间

## 项目状态

🔄 **当前阶段**：Phase 2.5 - 前后端集成

| 阶段 | 状态 | 说明 |
|------|------|------|
| 第一阶段 | ✅ 已完成 | Go 后端核心功能 |
| 第二阶段 | ✅ 已完成 | React 前端界面 + 完整集成 |
| 第2.5阶段 | 🔄 进行中 | 前后端集成（单一可执行文件） |
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

### 第一阶段（已完成）✅

- ✅ **Token 鉴权**（环境变量配置）
- ✅ Docker 容器管理（创建、启动、停止、删除、重置）
- ✅ 自定义脚本执行
- ✅ WebSSH 终端访问
- ✅ HTTP 端口转发（动态访问 + 端口标签）
- ✅ 数据持久化（工作空间配置自动恢复）

### 第二阶段（已完成）✅

- ✅ **React 前端界面**（Vite + TypeScript + Tailwind CSS）
- ✅ **工作空间可视化管理**（创建、删除、重置、状态监控）
- ✅ **Web 终端**（xterm.js + WebSocket 集成）
- ✅ **端口管理界面**（快捷访问、端口标签）
- ✅ **用户认证界面**（Token 登录、Cookie 会话管理）
- ✅ **响应式设计**（桌面、平板、移动端适配）
- ✅ **实时状态更新**（轮询机制）
- ✅ **错误处理与通知**（Toast 提示、全局错误处理）

### 第三阶段（计划）

- ⏳ GitHub 集成
- ⏳ AI Coding Agent 集成
- ⏳ VS Code Server 集成
- ⏳ 用户认证与权限管理

## 技术栈

### 后端

- **语言**：Go 1.25+
- **Web 框架**：Gin
- **Docker SDK**：github.com/docker/docker/client
- **WebSocket**：github.com/gorilla/websocket
- **反向代理**：net/http/httputil（标准库）

### 前端

- **框架**：React 18.3+ (函数组件 + Hooks)
- **构建工具**：Vite 7.2+
- **语言**：TypeScript 5.9+
- **样式**：Tailwind CSS 4.1+ (utility-first)
- **UI 组件**：shadcn UI (基于 Radix UI)
- **状态管理**：Jotai 2.15+ (原子化状态)
- **路由**：React Router DOM 7.9+
- **HTTP 客户端**：Axios 1.13+
- **终端模拟器**：xterm.js 5.5+ (WebGL 渲染)
- **通知组件**：Sonner 2.0+
- **图标库**：Lucide React

## 文档

### 后端文档
- [项目路线图](./PROJECT_ROADMAP.md) - 三个阶段的详细规划
- [第一阶段：后端实现](./docs/PHASE1_BACKEND.md) - 后端技术文档
- [任务拆分方案](./docs/PHASE1_TASK_BREAKDOWN.md) - 模块化开发和并行任务分配
- [API 规范](./docs/API_SPECIFICATION.md) - RESTful API 和 WebSocket 接口定义
- [后端增强方案](./docs/BACKEND_ENHANCEMENTS.md) - 端口标签、容器重置、数据持久化

### 前端文档
- [第二阶段：前端开发](./docs/PHASE2_FRONTEND.md) - 前端技术文档
- [前端任务拆分](./docs/PHASE2_TASK_BREAKDOWN.md) - 8个模块的详细任务拆分
- [Module 1-8 完成报告](./docs/archive/phase2-modules/) - 各模块的实现报告

### 集成文档
- [第2.5阶段：前后端集成](./docs/PHASE2.5_INTEGRATION.md) - 单一可执行文件集成方案

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

> **注意**：所有 API 都需要 Cookie 鉴权（浏览器）或查询参数鉴权（外部工具）

### 登录（设置 Cookie）

```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"token": "your-secret-token"}'
```

### 创建工作空间

```bash
# 使用 Cookie（浏览器自动发送）
curl -X POST http://localhost:3000/api/workspaces \
  -H "Content-Type: application/json" \
  -b cookies.txt \
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
```

### 访问终端

```javascript
// WebSocket 会自动发送 Cookie，也支持查询参数（备选）
const ws = new WebSocket('ws://localhost:3000/ws/terminal/{workspace-id}?token=your-secret-token');
ws.onmessage = (event) => console.log(event.data);
ws.send(JSON.stringify({type: 'input', data: 'ls -la\n'}));
```

### 访问容器内 HTTP 服务

```bash
# 浏览器访问（Cookie自动发送）：
http://localhost:3000/forward/{workspace-id}/8080/

# 外部工具访问（使用查询参数）：
curl "http://localhost:3000/forward/{workspace-id}/8080/?token=your-secret-token"
```

## 开发

### 环境要求

**后端**:
- Go 1.25+
- Docker
- Git

**前端**:
- Node.js 18+
- npm or pnpm

### 本地开发

#### 后端开发

```bash
# 克隆项目
git clone https://github.com/1PercentSync/vibox.git
cd vibox

# 安装依赖
go mod download

# 设置 API Token（必须）
export API_TOKEN=dev-token-123

# 运行后端
go run ./cmd/server

# 后端将运行在 http://localhost:3000
```

#### 前端开发

```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 前端将运行在 http://localhost:5173
# Vite 自动代理 API 请求到 http://localhost:3000
```

#### Mock 服务器（前端独立开发）

```bash
# 在 frontend 目录下

# 启动 Mock 服务器（端口 3000）
npm run mock

# 在另一个终端启动前端
npm run dev
```

#### 生产构建

```bash
# 构建前端
cd frontend
npm run build

# 输出到 frontend/dist/

# 预览生产构建
npm run preview

# 访问 http://localhost:4173
```

#### 完整集成（前端 + 后端）

```bash
# 1. 构建前端
cd frontend
npm run build

# 2. 将构建产物嵌入到 Go 后端
# TODO: 实现静态文件嵌入

# 3. 构建后端
cd ..
go build -o vibox ./cmd/server

# 4. 运行单一可执行文件
export API_TOKEN=your-secret-token
./vibox
```

### 开发进度

- **第一阶段 (后端)**：✅ 已完成 - 参见 [第一阶段开发计划](./docs/PHASE1_BACKEND.md#开发计划)
- **第二阶段 (前端)**：✅ 已完成 - 参见 [前端任务拆分](./docs/PHASE2_TASK_BREAKDOWN.md)
- **第2.5阶段 (集成)**：🔄 进行中 - 参见 [集成文档](./docs/PHASE2.5_INTEGRATION.md)
- **第三阶段**：⏳ 计划中

## 贡献

项目正在开发中，欢迎贡献！

## 许可证

待定

## 联系方式

- GitHub: https://github.com/1PercentSync/vibox
- Issue: https://github.com/1PercentSync/vibox/issues
