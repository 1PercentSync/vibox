# ViBox 技术选型与架构设计

## 技术调研总结

基于对现有技术的调研，我们分析了以下关键领域的解决方案。

### 1. 容器化开发环境（参考：DevContainer）

**DevContainer 工作原理**：
- 使用 `devcontainer.json` 文件定义开发环境配置
- IDE（如 VS Code）通过 Docker API 创建/连接容器
- 支持预安装工具、扩展、端口转发
- 工作区文件挂载或复制到容器内
- 2025年已成为主流的开发环境标准

**对 ViBox 的启示**：
- 可以借鉴配置文件驱动的方式
- 容器内提供完整的开发环境
- 支持自定义脚本初始化环境

### 2. Web 终端方案

**主流方案对比**：

| 方案 | 技术栈 | 优势 | 适用场景 |
|------|--------|------|---------|
| **WeTTY** | Node.js + xterm.js + WebSocket | 轻量、现代、活跃维护 | ✅ 推荐 |
| WebSSH2 | Node.js + xterm.js + socket.io | 功能丰富 | 备选 |
| Shellinabox | C + 自有前端 | 老牌稳定 | 不推荐（技术老旧） |

**xterm.js**：
- 完整的终端模拟器（JS实现）
- 支持所有现代浏览器
- 高性能、功能完整

### 3. Docker 容器管理

**方案对比**：

| 方案 | 语言 | 成熟度 | 适用场景 |
|------|------|--------|---------|
| **docker-py** (Python) | Python | 官方SDK，7.1.0 | ✅ 推荐（如果用Python） |
| **dockerode** (Node.js) | JavaScript | 社区维护，成熟稳定 | ✅ 推荐（如果用Node.js） |
| Docker API | REST API | 官方 | 直接调用 |
| Portainer | Go | 企业级GUI工具 | 不适合（我们要自己开发） |

### 4. 反向代理与端口转发

**方案对比**：

| 方案 | 特点 | 动态配置 | 学习曲线 |
|------|------|---------|---------|
| **Traefik** | 容器原生，自动服务发现 | ✅ 通过Docker标签 | 中等 |
| Nginx | 传统强大 | ❌ 需手动更新配置 | 低 |
| Caddy | 现代简洁 | 部分支持 | 低 |

**Traefik 优势**：
- 自动监听 Docker 事件
- 通过容器标签配置路由（无需重启）
- 原生支持动态服务发现
- 非常适合容器化环境

---

## MVP 技术选型建议

### 方案 A：Node.js 全栈（推荐）

**技术栈**：
```
前端：React/Vue + WebSocket
后端：Node.js + Express
Web终端：WeTTY (xterm.js)
容器管理：dockerode
反向代理：Traefik
容器基础：Ubuntu 22.04
```

**优势**：
- 前后端统一语言（JavaScript/TypeScript）
- dockerode 成熟稳定，社区活跃
- WeTTY 与 Node.js 生态集成良好
- 开发效率高

**架构图**：
```
┌──────────────────────────────────────────┐
│         浏览器 (Browser)                  │
│  ┌────────────┐  ┌──────────────────┐    │
│  │  前端 UI   │  │ xterm.js终端     │    │
│  └─────┬──────┘  └────────┬─────────┘    │
└────────┼───────────────────┼──────────────┘
         │ HTTP/WS           │ WebSocket
         ▼                   ▼
┌──────────────────────────────────────────┐
│      Node.js 后端服务 (Express)          │
│  ┌─────────┐  ┌────────────────────┐    │
│  │ API     │  │ WeTTY WebSocket    │    │
│  │ 服务    │  │ Server             │    │
│  └────┬────┘  └──────┬─────────────┘    │
│       │              │                   │
│  ┌────▼──────────────▼────────┐         │
│  │    dockerode (Docker SDK)  │         │
│  └────────────┬────────────────┘         │
└───────────────┼──────────────────────────┘
                │ Docker API
                ▼
┌──────────────────────────────────────────┐
│         Docker Engine                     │
│  ┌────────────────────────────────────┐  │
│  │  Traefik (反向代理容器)             │  │
│  │  - 监听Docker事件                   │  │
│  │  - 自动配置路由                     │  │
│  └────────────┬───────────────────────┘  │
│               │                           │
│  ┌────────────▼───────────────────────┐  │
│  │  工作空间容器 (动态创建)            │  │
│  │  - Ubuntu 22.04                     │  │
│  │  - 开发工具                         │  │
│  │  - 用户HTTP服务 (8080, 3000...)    │  │
│  └─────────────────────────────────────┘  │
└──────────────────────────────────────────┘
```

### 方案 B：Python 后端（备选）

**技术栈**：
```
前端：React/Vue
后端：Python (FastAPI/Flask)
Web终端：WeTTY (独立服务) 或集成 xterm.js
容器管理：docker-py (官方SDK)
反向代理：Traefik
```

**优势**：
- docker-py 是官方SDK
- Python 生态丰富
- FastAPI 性能好，异步支持

**劣势**：
- WeTTY 是Node.js项目，需要独立部署或重写
- 前后端语言分离

---

## 推荐架构：Node.js 方案

### 核心组件

#### 1. 前端 (Frontend)
```
技术：React + TypeScript + Vite
功能：
- 工作空间管理界面
- Web 终端集成（xterm.js）
- 端口转发管理
- 实时状态显示
```

#### 2. 后端 (Backend)
```
技术：Node.js + Express + TypeScript
依赖：
- dockerode: Docker容器管理
- ws: WebSocket支持（Web终端）
- node-pty: 伪终端（用于终端会话）

核心功能：
- RESTful API（工作空间CRUD）
- WebSocket服务（Web终端）
- Docker容器编排
- 脚本执行管理
```

#### 3. 反向代理 (Traefik)
```
部署：Docker容器
配置：
- 监听Docker socket
- 基于标签自动路由
- 动态服务发现

路由规则示例：
/forward/{workspace-id}/{port} -> 容器内端口
```

#### 4. 工作空间容器
```
基础镜像：ubuntu:22.04
预装工具：
- bash, curl, wget, git
- 常用开发工具

启动流程：
1. 创建容器（使用dockerode）
2. 复制用户脚本到容器
3. 执行初始化脚本（按顺序）
4. 添加Traefik标签（用于端口转发）
5. 启动SSH服务（供WeTTY连接）
```

---

## 数据模型设计

### 工作空间 (Workspace)

```typescript
interface Workspace {
  id: string;                    // 唯一标识
  name: string;                  // 工作空间名称
  containerId: string;           // Docker容器ID
  status: 'creating' | 'running' | 'stopped' | 'error';
  createdAt: Date;

  // 配置
  config: {
    scripts: Script[];           // 初始化脚本
    exposedPorts: ExposedPort[]; // 暴露的端口
  };
}

interface Script {
  id: string;
  name: string;
  content: string;              // 脚本内容
  order: number;                // 执行顺序
}

interface ExposedPort {
  containerPort: number;        // 容器内端口
  enabled: boolean;             // 是否启用
  publicUrl?: string;           // 公开访问URL
}
```

---

## API 设计

### RESTful API

```typescript
// 工作空间管理
POST   /api/workspaces              // 创建工作空间
GET    /api/workspaces              // 列出所有工作空间
GET    /api/workspaces/:id          // 获取工作空间详情
DELETE /api/workspaces/:id          // 删除工作空间
POST   /api/workspaces/:id/start    // 启动工作空间
POST   /api/workspaces/:id/stop     // 停止工作空间

// 端口管理
POST   /api/workspaces/:id/ports    // 启用端口转发
DELETE /api/workspaces/:id/ports/:port  // 禁用端口转发
GET    /api/workspaces/:id/ports    // 列出已转发端口

// 脚本管理
POST   /api/scripts                 // 上传脚本
GET    /api/scripts                 // 列出脚本
DELETE /api/scripts/:id             // 删除脚本
```

### WebSocket API

```typescript
// Web 终端
WS /ws/terminal/:workspaceId
- 消息格式：{ type: 'data' | 'resize', data: string | {cols, rows} }
```

---

## 实现步骤

### Phase 1: 基础架构搭建
1. 初始化项目结构（monorepo: frontend + backend）
2. 配置开发环境（TypeScript, ESLint, etc）
3. 搭建基础 Express 服务
4. 集成 dockerode

### Phase 2: 容器管理
1. 实现容器创建 API
2. 实现脚本注入和执行
3. 实现容器生命周期管理（启动/停止/删除）
4. 容器状态监控

### Phase 3: Web 终端
1. 集成 xterm.js 到前端
2. 实现 WebSocket 终端服务
3. 使用 node-pty 连接到容器
4. 实现终端会话管理

### Phase 4: 端口转发
1. 部署 Traefik 容器
2. 实现动态添加 Traefik 标签
3. 生成端口转发 URL
4. 前端显示可访问链接

### Phase 5: 前端界面
1. 工作空间列表页面
2. 工作空间创建表单（脚本上传）
3. 工作空间详情页（终端 + 端口管理）
4. 实时状态更新

### Phase 6: 测试与优化
1. 单元测试
2. 集成测试
3. 性能优化
4. 错误处理完善

---

## 部署方案

### 开发环境
```yaml
# docker-compose.yml
services:
  backend:
    build: ./backend
    ports:
      - "3000:3000"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock

  frontend:
    build: ./frontend
    ports:
      - "5173:5173"

  traefik:
    image: traefik:v2.10
    ports:
      - "80:80"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```

### 生产环境
- 所有服务容器化
- 使用 Docker Compose 或 Kubernetes 编排
- 添加持久化存储（配置、脚本）
- 配置日志收集
- 添加监控和告警

---

## 技术风险与挑战

### 1. Docker Socket 访问
**风险**：后端需要访问 Docker socket，安全风险较高
**解决**：
- 限制容器权限
- 考虑使用 Docker API over TCP（需要 TLS）
- 生产环境使用专门的容器编排工具

### 2. Web 终端性能
**风险**：多个并发终端会话可能影响性能
**解决**：
- 限制每个用户的终端会话数
- 实现会话超时机制
- 使用 WebSocket 连接池

### 3. 端口冲突
**风险**：多个容器可能尝试使用相同的主机端口
**解决**：
- 不直接映射主机端口
- 完全通过 Traefik 代理访问
- Traefik 基于域名/路径路由，避免端口冲突

### 4. 容器资源管理
**风险**：用户可能创建过多容器，耗尽资源
**解决**：
- 设置容器资源限制（CPU、内存）
- 限制用户可创建的容器数量
- 实现自动清理机制（闲置容器）

---

---

## 最终技术选型 (2025-11-10)

### 后端：Go 语言 ✅

**确定使用 Go 开发后端**，原因：
- 开发者对 Go 感兴趣，有学习动力
- 所有需要的库都很成熟（Docker SDK、PTY、WebSocket、反向代理）
- 部署简单（单个二进制文件）
- 性能优秀，天然支持高并发
- 标准库强大，反向代理直接用 `net/http/httputil`

**技术栈**：
```
语言: Go 1.21+
Web框架: Gin (简单易用)
依赖:
  - github.com/docker/docker/client  (Docker 容器管理，官方SDK)
  - github.com/creack/pty            (伪终端 PTY 支持)
  - github.com/gorilla/websocket     (WebSocket 支持)
  - github.com/gin-gonic/gin         (Web 框架)
  - net/http/httputil                (反向代理，标准库)

部署: Docker + Docker Compose
唯一对外端口: 3000 (适配用户 Caddy 反向代理)
```

### 前端：待定 ⏳

前端技术栈暂时不确定，以下是待选方案：

**方案 1: React + TypeScript**
```
框架: React 18
语言: TypeScript
构建工具: Vite
终端: xterm.js
UI库: Ant Design / Tailwind CSS (待定)
```

**方案 2: Vue + TypeScript**
```
框架: Vue 3
语言: TypeScript
构建工具: Vite
终端: xterm.js
UI库: Element Plus / Naive UI (待定)
```

**方案 3: 原生 HTML + JavaScript**
```
更简单，但功能可能受限
```

前端方案将在后端 MVP 完成后再确定。

---

## 部署架构（最终版）

```
用户浏览器
    ↓
用户的 Caddy (domain.com)
    ↓
反向代理到 localhost:3000
    ↓
┌─────────────────────────────────────────┐
│  Go 后端服务 (唯一对外端口: 3000)        │
│                                          │
│  路由规则：                               │
│  /           → 前端静态文件 (内嵌)        │
│  /api/*      → RESTful API               │
│  /ws/*       → WebSocket (Web 终端)      │
│  /forward/:workspace/:port/* → 反向代理   │
│                                          │
│  ┌────────────────────────────────┐     │
│  │  httputil.ReverseProxy         │     │
│  │  (标准库动态反向代理)           │     │
│  └────────────┬───────────────────┘     │
│               │                          │
│  ┌────────────▼───────────────────┐     │
│  │  Docker SDK                    │     │
│  │  github.com/docker/docker/     │     │
│  │  client                        │     │
│  └────────────┬───────────────────┘     │
└───────────────┼─────────────────────────┘
                │ Docker API
                ↓
┌─────────────────────────────────────────┐
│         Docker Engine                    │
│  ┌─────────────────────────────────┐    │
│  │  工作空间容器 (动态创建)         │    │
│  │  - Ubuntu 22.04                 │    │
│  │  - 开发工具                     │    │
│  │  - 用户HTTP服务                 │    │
│  │  - 不暴露端口到宿主机           │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

**注意**：
- 不再使用 Traefik，反向代理功能由 Go 后端直接实现
- 工作空间容器不暴露任何端口到宿主机
- 所有访问通过后端的反向代理转发到容器内部 IP

---

## 下一步

详细设计文档参见：
- [WebSSH 实现原理](./docs/WEBSSH_PRINCIPLE.md)（待创建）
- [Go 后端架构设计](./docs/BACKEND_ARCHITECTURE.md)（待创建）
- [开发计划](./docs/DEVELOPMENT_PLAN.md)（待创建）
