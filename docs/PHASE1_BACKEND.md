# 第一阶段：Go 后端实现

> **当前阶段目标**：实现完整的 Go 后端服务，提供所有核心 API 和功能

---

## 目录

1. [技术栈](#技术栈)
2. [WebSSH 实现原理](#webssh-实现原理)
3. [项目架构](#项目架构)
4. [开发计划](#开发计划)
5. [部署方案](#部署方案)

---

## 技术栈

### 核心依赖

```go
require (
    github.com/gin-gonic/gin v1.9.1           // Web 框架
    github.com/docker/docker v24.0.0          // Docker SDK（官方）
    github.com/gorilla/websocket v1.5.1       // WebSocket
    github.com/google/uuid v1.5.0             // UUID 生成
)
```

### 标准库

```go
import (
    "net/http/httputil"  // 反向代理
    "context"            // 上下文管理
    "io"                 // 流式 I/O
)
```

### 为什么选择这些技术？

- **Gin**：轻量级 Web 框架，性能优秀，社区活跃
- **Docker SDK**：官方 SDK，2025 年仍在持续更新
- **gorilla/websocket**：成熟稳定的 WebSocket 实现
- **httputil**：标准库反向代理，无需第三方依赖

---

## WebSSH 实现原理

### 核心概念

#### 1. 伪终端（PTY）

PTY 是一对虚拟字符设备：
- **Master 端**：Go 后端读写
- **Slave 端**：Shell 进程读写，认为在真实终端中

**为什么需要 PTY？**
- 提供交互式提示符
- 支持终端控制序列（颜色、光标）
- 支持窗口大小调整
- 交互式程序（vim、top）可正常工作

#### 2. Docker Exec API vs SSH

我们使用 **Docker Exec API** 而不是 SSH：

| 方法 | 优势 | 劣势 |
|------|------|------|
| **Docker Exec** | 无需 SSH 服务，更轻量 | 需要访问 Docker Socket |
| SSH | 标准协议 | 需要在容器内安装 SSH |

#### 3. 数据流程

```
用户浏览器
    ↓ (用户输入 "ls -la")
xterm.js 捕获键盘
    ↓ WebSocket 发送
    {"type": "input", "data": "ls -la\n"}
    ↓
Go 后端 WebSocket Handler
    ↓ 写入 Docker Exec 连接
Docker 容器内 /bin/bash
    ↓ 执行命令
    "total 48\ndrwxr-xr-x..."
    ↓ 通过 Docker Exec 返回
Go 后端读取输出
    ↓ WebSocket 发送
    {"type": "output", "data": "..."}
    ↓
xterm.js 渲染
    ↓
显示在浏览器
```

### 实现步骤

**Step 1: 前端 WebSocket 连接**
```javascript
const ws = new WebSocket('ws://domain.com/ws/terminal/workspace-123');
const term = new Terminal();

term.onData(data => ws.send(JSON.stringify({type: 'input', data})));
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'output') term.write(msg.data);
};
```

**Step 2: Go 后端处理**
```go
// 1. 升级为 WebSocket
ws, _ := upgrader.Upgrade(c.Writer, c.Request, nil)

// 2. 在容器中创建 Exec 实例
exec, _ := dockerClient.ContainerExecCreate(ctx, containerID, types.ExecConfig{
    Cmd:  []string{"/bin/bash"},
    Tty:  true,  // 关键！
    AttachStdin:  true,
    AttachStdout: true,
})

// 3. 连接到 Exec
hijacked, _ := dockerClient.ContainerExecAttach(ctx, exec.ID, ...)

// 4. 双向数据传输
go WebSocketToExec(ws, hijacked)  // 用户输入 → 容器
go ExecToWebSocket(hijacked, ws)  // 容器输出 → 用户
```

**Step 3: 容器执行**
- Docker 在容器内启动 `/bin/bash`
- Bash 检测到 TTY，显示提示符
- 接收命令，执行，返回输出

---

## 项目架构

### 目录结构

```
vibox/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── api/
│   │   ├── handler/             # HTTP/WebSocket 处理器
│   │   │   ├── workspace.go     # 工作空间 API
│   │   │   ├── terminal.go      # WebSocket 终端
│   │   │   └── proxy.go         # 端口转发
│   │   ├── middleware/          # 中间件
│   │   │   ├── cors.go
│   │   │   ├── logger.go
│   │   │   └── recovery.go
│   │   └── router.go            # 路由配置
│   ├── domain/                  # 领域模型
│   │   └── workspace.go
│   ├── service/                 # 业务逻辑
│   │   ├── docker.go            # Docker 操作
│   │   ├── workspace.go         # 工作空间管理
│   │   ├── terminal.go          # 终端会话
│   │   └── proxy.go             # 反向代理
│   ├── repository/              # 数据持久化
│   │   └── workspace.go
│   └── config/
│       └── config.go            # 配置管理
├── pkg/
│   └── utils/
│       ├── id.go                # ID 生成
│       └── logger.go            # 日志工具
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── README.md
```

### 核心模块

#### 1. Domain（领域模型）

```go
type Workspace struct {
    ID          string          `json:"id"`
    Name        string          `json:"name"`
    ContainerID string          `json:"container_id"`
    Status      WorkspaceStatus `json:"status"` // creating/running/stopped/error
    CreatedAt   time.Time       `json:"created_at"`
    Config      WorkspaceConfig `json:"config"`
}

type WorkspaceConfig struct {
    Image        string         `json:"image"`         // 默认 ubuntu:22.04
    Scripts      []Script       `json:"scripts"`       // 初始化脚本
    ExposedPorts []ExposedPort  `json:"exposed_ports"` // 暴露的端口
}
```

#### 2. Service（业务逻辑）

**DockerService** - Docker 操作封装
```go
func (s *DockerService) CreateContainer(ctx, config) (containerID, error)
func (s *DockerService) StartContainer(ctx, containerID) error
func (s *DockerService) StopContainer(ctx, containerID) error
func (s *DockerService) RemoveContainer(ctx, containerID) error
func (s *DockerService) GetContainerIP(ctx, containerID) (ip, error)
```

**WorkspaceService** - 工作空间管理
```go
func (s *WorkspaceService) CreateWorkspace(ctx, req) (*Workspace, error)
func (s *WorkspaceService) GetWorkspace(id) (*Workspace, error)
func (s *WorkspaceService) ListWorkspaces() ([]*Workspace, error)
func (s *WorkspaceService) DeleteWorkspace(ctx, id) error
```

**TerminalService** - 终端会话
```go
func (s *TerminalService) CreateSession(ws, containerID) (*TerminalSession, error)
// TerminalSession 内部处理双向数据传输
```

**ProxyService** - 端口转发
```go
func (s *ProxyService) ProxyRequest(w, r, containerID, port) error
// 使用 httputil.ReverseProxy 实现
```

#### 3. API（路由）

```go
// 工作空间管理
POST   /api/workspaces              // 创建工作空间
GET    /api/workspaces              // 列出工作空间
GET    /api/workspaces/:id          // 获取工作空间
DELETE /api/workspaces/:id          // 删除工作空间

// WebSocket 终端
GET    /ws/terminal/:id             // 连接到终端

// 端口转发
ANY    /forward/:id/:port/*path     // 转发到容器端口
```

---

## 开发计划

### 时间规划（15-21 天）

| 阶段 | 任务 | 时间 |
|------|------|------|
| **Phase 0** | 环境准备 | 1 天 |
| **Phase 1** | 基础架构（配置、Docker、路由） | 2-3 天 |
| **Phase 2** | 工作空间管理（CRUD + 脚本） | 3-4 天 |
| **Phase 3** | WebSSH 终端（核心功能）| 4-5 天 |
| **Phase 4** | HTTP 端口转发 | 2-3 天 |
| **Phase 5** | 完善优化 | 2-3 天 |
| **Phase 6** | 容器化部署 | 1-2 天 |

### Phase 0: 环境准备（1 天）

**任务清单**：
- [ ] 安装 Go 1.21+
- [ ] 安装 Docker
- [ ] 初始化项目：`go mod init vibox`
- [ ] 创建目录结构
- [ ] 配置 `.gitignore`
- [ ] 安装依赖：
  ```bash
  go get github.com/gin-gonic/gin
  go get github.com/docker/docker
  go get github.com/gorilla/websocket
  go get github.com/google/uuid
  ```

**验收标准**：
- 项目结构清晰
- 可以运行 `go build ./cmd/server`

### Phase 1: 基础架构（2-3 天）

**任务清单**：
- [ ] 实现配置管理（`internal/config/config.go`）
- [ ] 实现工具函数（ID 生成、日志）
- [ ] 实现领域模型（`internal/domain/workspace.go`）
- [ ] 实现 DockerService 基础功能
- [ ] 实现 Repository 接口（内存存储）
- [ ] 实现中间件（Logger、Recovery、CORS）
- [ ] 实现基础路由
- [ ] 实现 `main.go` 入口

**验收标准**：
- 服务可以启动在 `:3000`
- 可以通过 Docker SDK 创建/删除容器
- 基础中间件工作正常

### Phase 2: 工作空间管理（3-4 天）

**任务清单**：
- [ ] 实现 WorkspaceService 完整逻辑
- [ ] 实现工作空间 CRUD API
- [ ] 实现脚本复制到容器
- [ ] 实现脚本按顺序执行
- [ ] 实现状态管理（creating → running/error）
- [ ] 实现容器健康检查

**测试**：
```bash
# 创建工作空间
curl -X POST http://localhost:3000/api/workspaces \
  -H "Content-Type: application/json" \
  -d '{"name": "test", "scripts": [...]}'

# 列出工作空间
curl http://localhost:3000/api/workspaces

# 删除工作空间
curl -X DELETE http://localhost:3000/api/workspaces/{id}
```

**验收标准**：
- 可以创建工作空间，容器自动启动
- 脚本按顺序执行成功
- 可以查询和删除工作空间

### Phase 3: WebSSH 终端（4-5 天）

**任务清单**：
- [ ] 实现 TerminalService
- [ ] 实现 WebSocket 升级
- [ ] 实现 Docker Exec 创建和 Attach
- [ ] 实现双向数据传输（goroutine）
- [ ] 实现终端 resize 支持
- [ ] 实现会话清理机制

**消息协议**：
```json
// 输入
{"type": "input", "data": "ls\n"}

// 输出
{"type": "output", "data": "file1 file2\n"}

// 调整大小
{"type": "resize", "cols": 80, "rows": 24}
```

**测试工具**：
```bash
# 使用 websocat 测试
websocat ws://localhost:3000/ws/terminal/{workspace-id}
```

**验收标准**：
- WebSocket 连接成功
- 可以执行命令并看到输出
- 支持交互式程序（如 vim）
- 终端大小调整正常
- 连接断开后资源清理

### Phase 4: HTTP 端口转发（2-3 天）

**任务清单**：
- [ ] 实现 ProxyService
- [ ] 使用 `httputil.ReverseProxy`
- [ ] 实现路径重写
- [ ] 实现错误处理

**测试**：
```bash
# 在容器内启动 HTTP 服务
docker exec {container} python3 -m http.server 8080

# 通过前端访问
curl http://localhost:3000/forward/{workspace-id}/8080/
```

**验收标准**：
- 可以访问容器内 HTTP 服务
- 路径正确转发
- POST/PUT 等请求正常

### Phase 5: 完善与优化（2-3 天）

**任务清单**：
- [ ] 统一错误处理
- [ ] 结构化日志（可选：使用 logrus/zap）
- [ ] 资源限制（容器 CPU/内存）
- [ ] 并发会话限制
- [ ] CORS 配置优化
- [ ] WebSocket Origin 检查
- [ ] API 文档（可选：Swagger）

**验收标准**：
- 错误信息清晰
- 日志完整可追踪
- 资源使用合理

### Phase 6: 容器化部署（1-2 天）

**任务清单**：
- [ ] 编写 Dockerfile（多阶段构建）
- [ ] 编写 docker-compose.yml
- [ ] 环境变量配置
- [ ] 测试容器部署

**Dockerfile 示例**：
```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o vibox ./cmd/server

FROM ubuntu:22.04
COPY --from=builder /app/vibox /usr/local/bin/
EXPOSE 3000
CMD ["vibox"]
```

**docker-compose.yml**：
```yaml
version: '3.8'
services:
  vibox:
    build: .
    ports:
      - "3000:3000"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DEFAULT_IMAGE=ubuntu:22.04
```

**验收标准**：
- 镜像构建成功
- 容器启动正常
- 所有功能正常工作

---

## 部署方案

### 开发环境

```bash
# 直接运行
go run ./cmd/server

# 或使用 docker-compose
docker-compose up
```

### 生产环境

```bash
# 构建镜像
docker build -t vibox:latest .

# 运行容器
docker run -d \
  -p 3000:3000 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --name vibox \
  vibox:latest
```

### 用户访问

```
用户配置 Caddy:

# Caddyfile
domain.com {
    reverse_proxy localhost:3000
}
```

用户访问 `https://domain.com` 即可使用 ViBox。

---

## 技术要点

### 1. Docker Socket 安全

**风险**：后端需要访问 `/var/run/docker.sock`，具有很高权限

**缓解措施**：
- 限制容器资源（CPU、内存）
- 设置容器网络隔离
- 考虑后续使用 Docker API over TCP + TLS

### 2. WebSocket 性能

**优化**：
- 限制每个用户的并发会话数
- 实现会话超时机制
- 使用缓冲读写

### 3. 容器资源管理

**配置**：
```go
Resources: container.Resources{
    Memory:   512 * 1024 * 1024,  // 512MB
    NanoCPUs: 1000000000,          // 1 CPU
}
```

### 4. 错误处理

所有 API 返回统一格式：
```json
{
  "error": "错误描述",
  "code": "ERROR_CODE"
}
```

---

## 参考资源

### Go 学习
- [Go 官方文档](https://go.dev/doc/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)

### Docker SDK
- [Docker SDK for Go](https://docs.docker.com/engine/api/sdk/)
- [GitHub: docker/docker](https://github.com/moby/moby)

### WebSocket
- [gorilla/websocket](https://github.com/gorilla/websocket)

### 相关项目
- [Portainer](https://github.com/portainer/portainer) - Docker 管理
- [Wetty](https://github.com/butlerx/wetty) - Web 终端
- [ttyd](https://github.com/tsl0922/ttyd) - 终端共享

---

## 成功标准

第一阶段完成后，应该能够：

- ✅ 通过 API 创建工作空间
- ✅ 容器自动启动并执行初始化脚本
- ✅ 通过 WebSocket 访问容器终端
- ✅ 通过 URL 访问容器内 HTTP 服务
- ✅ 删除工作空间及容器
- ✅ 使用 Docker Compose 一键部署

**下一步**：进入[第二阶段](../PROJECT_ROADMAP.md#第二阶段前端界面--mvp-集成--待定)，开发前端界面。
