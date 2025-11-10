# 第一阶段后端开发任务拆分

> **目标**：将后端开发拆分为独立模块，支持多个 agent 并行开发

---

## 目录

1. [模块依赖关系图](#模块依赖关系图)
2. [并行开发策略](#并行开发策略)
3. [模块详细说明](#模块详细说明)
4. [接口定义](#接口定义)
5. [开发顺序建议](#开发顺序建议)

---

## 模块依赖关系图

```
┌─────────────────────────────────────────────────────────┐
│  Layer 1: 基础设施层 (Foundation)                        │
│  - Config 配置管理                                       │
│  - Logger 日志工具                                       │
│  - Utils 工具函数                                        │
│  - Middleware (Auth, CORS, Logger, Recovery)            │
└─────────────────┬───────────────────────────────────────┘
                  │
    ┌─────────────┴─────────────┐
    │                           │
┌───▼────────────────┐   ┌─────▼─────────────────────────┐
│  Layer 2: 核心服务  │   │  Layer 2: 数据层               │
│  - DockerService   │   │  - Domain Models              │
│                    │   │  - Repository (Memory)        │
└───┬────────────────┘   └─────┬─────────────────────────┘
    │                           │
    │    ┌──────────────────────┴────────────────────────┐
    │    │                      │                        │
┌───▼────▼──────────┐  ┌───────▼───────┐  ┌────────────▼─────────┐
│  Layer 3: 业务服务 │  │  Layer 3:     │  │  Layer 3:            │
│  - WorkspaceService│  │  TerminalSvc  │  │  ProxyService        │
└───┬────────────────┘  └───────┬───────┘  └────────────┬─────────┘
    │                           │                        │
    └───────────────┬───────────┴────────────────────────┘
                    │
            ┌───────▼────────┐
            │  Layer 4: API  │
            │  - Router      │
            │  - Handlers    │
            └───────┬────────┘
                    │
            ┌───────▼────────┐
            │  Layer 5:      │
            │  - Deployment  │
            │  - CI/CD       │
            └────────────────┘
```

**依赖说明**：
- **Layer 1** → 所有其他层都依赖（基础设施）
- **Layer 2** → Layer 3, Layer 4 依赖
- **Layer 3** → Layer 4 依赖
- **Layer 4** → Layer 5 依赖

---

## 并行开发策略

### 🚀 开发轮次

| 轮次 | 模块 | 预计时间 | 可并行 Agent 数 |
|------|------|----------|----------------|
| **Round 1** | Module 1: 基础设施层 | 1-2 天 | 1-2 个 |
| **Round 2** | Module 2: Docker 服务 + Module 3a: 数据层 | 2-3 天 | 2 个 |
| **Round 3** | Module 3b: 工作空间服务 + Module 4: 终端服务 + Module 5: 代理服务 | 4-6 天 | 3 个 |
| **Round 4** | Module 6: API 层 | 2-3 天 | 2 个 |
| **Round 5** | Module 7: 部署和 CI/CD | 1-2 天 | 1 个 |

**总预计时间**：10-16 天（与原计划 15-21 天一致，通过并行加速）

---

## 模块详细说明

### Module 1: 基础设施层 (Foundation)

**负责人**：Agent 1
**预计时间**：1-2 天
**优先级**：🔴 最高（所有模块依赖）

#### 📦 包含组件

1. **Config 配置管理** (`internal/config/config.go`)
2. **Logger 日志工具** (`pkg/utils/logger.go`)
3. **Utils 工具函数** (`pkg/utils/id.go`)
4. **Middleware** (`internal/api/middleware/`)
   - `auth.go` - Token 鉴权
   - `cors.go` - CORS 处理
   - `logger.go` - 请求日志
   - `recovery.go` - Panic 恢复

#### 📋 任务清单

- [ ] 实现 Config 配置管理
  - [ ] 从环境变量读取配置
  - [ ] 支持 `API_TOKEN`, `PORT`, `DOCKER_HOST`, `DEFAULT_IMAGE`
  - [ ] 启动时验证必需配置（API_TOKEN 不能为空）
- [ ] 实现 Logger 工具
  - [ ] 结构化日志（建议使用标准库 `log/slog`）
  - [ ] 日志级别：DEBUG, INFO, WARN, ERROR
- [ ] 实现 Utils 工具
  - [ ] UUID 生成（使用 `github.com/google/uuid`）
  - [ ] ID 验证函数
- [ ] 实现 Middleware
  - [ ] Auth 中间件（支持 Header 和查询参数）
  - [ ] CORS 中间件（配置允许的 origin）
  - [ ] Logger 中间件（记录请求日志）
  - [ ] Recovery 中间件（捕获 panic）

#### 🔌 对外接口

**Config 接口**：
```go
type Config struct {
    Port         string
    APIToken     string
    DockerHost   string
    DefaultImage string
    MemoryLimit  int64
    CPULimit     int64
}

func Load() *Config
```

**Logger 接口**：
```go
func Debug(msg string, args ...any)
func Info(msg string, args ...any)
func Warn(msg string, args ...any)
func Error(msg string, args ...any)
```

**Utils 接口**：
```go
func GenerateID() string
func ValidateID(id string) error
```

**Middleware 接口**：
```go
func AuthMiddleware(requiredToken string) gin.HandlerFunc
func CORSMiddleware() gin.HandlerFunc
func LoggerMiddleware() gin.HandlerFunc
func RecoveryMiddleware() gin.HandlerFunc
```

#### ✅ 验收标准

- [ ] 配置从环境变量正确读取
- [ ] API_TOKEN 未设置时程序拒绝启动
- [ ] 日志正常输出到 stdout
- [ ] Auth 中间件正确拦截未授权请求
- [ ] 所有中间件通过单元测试

#### 📚 依赖

- Go 标准库
- `github.com/gin-gonic/gin`
- `github.com/google/uuid`

---

### Module 2: Docker 服务层 (Docker Service)

**负责人**：Agent 2
**预计时间**：2-3 天
**优先级**：🔴 高（核心服务）
**依赖**：Module 1

#### 📦 包含组件

1. **DockerService** (`internal/service/docker.go`)

#### 📋 任务清单

- [ ] 初始化 Docker 客户端
- [ ] 实现容器生命周期管理
  - [ ] `CreateContainer(ctx, config) (containerID, error)`
  - [ ] `StartContainer(ctx, containerID) error`
  - [ ] `StopContainer(ctx, containerID, timeout) error`
  - [ ] `RemoveContainer(ctx, containerID) error`
- [ ] 实现容器信息查询
  - [ ] `GetContainerIP(ctx, containerID) (ip, error)`
  - [ ] `GetContainerStatus(ctx, containerID) (status, error)`
  - [ ] `InspectContainer(ctx, containerID) (*types.ContainerJSON, error)`
- [ ] 实现脚本执行
  - [ ] `ExecCommand(ctx, containerID, cmd []string) (output, error)`
  - [ ] `CopyToContainer(ctx, containerID, path, content) error`
- [ ] 配置资源限制（CPU、内存）
- [ ] 错误处理和日志

#### 🔌 对外接口

```go
type DockerService struct {
    client *client.Client
    config *config.Config
}

func NewDockerService(cfg *config.Config) (*DockerService, error)

// 容器生命周期
func (s *DockerService) CreateContainer(ctx context.Context, cfg ContainerConfig) (string, error)
func (s *DockerService) StartContainer(ctx context.Context, containerID string) error
func (s *DockerService) StopContainer(ctx context.Context, containerID string, timeout int) error
func (s *DockerService) RemoveContainer(ctx context.Context, containerID string) error

// 容器信息
func (s *DockerService) GetContainerIP(ctx context.Context, containerID string) (string, error)
func (s *DockerService) GetContainerStatus(ctx context.Context, containerID string) (string, error)

// 脚本执行
func (s *DockerService) ExecCommand(ctx context.Context, containerID string, cmd []string) (string, error)
func (s *DockerService) CopyToContainer(ctx context.Context, containerID string, path string, content []byte) error

// 类型定义
type ContainerConfig struct {
    Image        string
    Name         string
    MemoryLimit  int64
    CPULimit     int64
}
```

#### ✅ 验收标准

- [ ] 可以成功创建、启动、停止、删除容器
- [ ] 资源限制正确应用
- [ ] 可以执行命令并获取输出
- [ ] 可以复制文件到容器
- [ ] 错误情况正确处理
- [ ] 通过单元测试（使用 testcontainers-go 或 mock）

#### 📚 依赖

- Module 1 (Config, Logger)
- `github.com/docker/docker/client`
- `github.com/docker/docker/api/types`

---

### Module 3a: 数据层 (Data Layer)

**负责人**：Agent 3
**预计时间**：1 天
**优先级**：🟡 中（可与 Module 2 并行）
**依赖**：Module 1

#### 📦 包含组件

1. **Domain 模型** (`internal/domain/workspace.go`)
2. **Repository 接口和实现** (`internal/repository/workspace.go`)

#### 📋 任务清单

- [ ] 定义 Domain 模型
  - [ ] `Workspace` 结构体
  - [ ] `WorkspaceConfig` 结构体
  - [ ] `Script` 结构体
  - [ ] `WorkspaceStatus` 枚举
- [ ] 定义 Repository 接口
- [ ] 实现内存存储（使用 `sync.Map` 或 map + mutex）
- [ ] 实现 CRUD 操作

#### 🔌 对外接口

**Domain 模型**：
```go
type WorkspaceStatus string

const (
    StatusCreating WorkspaceStatus = "creating"
    StatusRunning  WorkspaceStatus = "running"
    StatusStopped  WorkspaceStatus = "stopped"
    StatusError    WorkspaceStatus = "error"
)

type Workspace struct {
    ID          string          `json:"id"`
    Name        string          `json:"name"`
    ContainerID string          `json:"container_id"`
    Status      WorkspaceStatus `json:"status"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
    Config      WorkspaceConfig `json:"config"`
    Error       string          `json:"error,omitempty"`
}

type WorkspaceConfig struct {
    Image        string   `json:"image"`
    Scripts      []Script `json:"scripts,omitempty"`
}

type Script struct {
    Name    string `json:"name"`
    Content string `json:"content"`
    Order   int    `json:"order"`
}
```

**Repository 接口**：
```go
type WorkspaceRepository interface {
    Create(ws *domain.Workspace) error
    Get(id string) (*domain.Workspace, error)
    List() ([]*domain.Workspace, error)
    Update(ws *domain.Workspace) error
    Delete(id string) error
}

type MemoryRepository struct {
    mu    sync.RWMutex
    store map[string]*domain.Workspace
}

func NewMemoryRepository() *MemoryRepository
```

#### ✅ 验收标准

- [ ] Domain 模型定义完整
- [ ] Repository 接口定义清晰
- [ ] 内存存储实现线程安全
- [ ] CRUD 操作正常工作
- [ ] 通过单元测试

#### 📚 依赖

- Module 1 (Logger)
- Go 标准库 (`sync`, `time`)

---

### Module 3b: 工作空间服务 (Workspace Service)

**负责人**：Agent 3（继续）
**预计时间**：2-3 天
**优先级**：🟡 中
**依赖**：Module 1, Module 2, Module 3a

#### 📦 包含组件

1. **WorkspaceService** (`internal/service/workspace.go`)

#### 📋 任务清单

- [ ] 实现工作空间创建流程
  - [ ] 生成 workspace ID
  - [ ] 创建 Docker 容器
  - [ ] 启动容器
  - [ ] 执行初始化脚本（按 order 排序）
  - [ ] 更新状态
- [ ] 实现工作空间查询
- [ ] 实现工作空间列表
- [ ] 实现工作空间删除
- [ ] 脚本执行逻辑（按顺序，失败停止）
- [ ] 状态管理（creating → running/error）
- [ ] 错误处理和回滚

#### 🔌 对外接口

```go
type WorkspaceService struct {
    dockerSvc *DockerService
    repo      repository.WorkspaceRepository
    config    *config.Config
}

func NewWorkspaceService(dockerSvc *DockerService, repo repository.WorkspaceRepository, cfg *config.Config) *WorkspaceService

// 工作空间管理
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req CreateWorkspaceRequest) (*domain.Workspace, error)
func (s *WorkspaceService) GetWorkspace(id string) (*domain.Workspace, error)
func (s *WorkspaceService) ListWorkspaces() ([]*domain.Workspace, error)
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id string) error

// 辅助方法
func (s *WorkspaceService) executeScripts(ctx context.Context, containerID string, scripts []domain.Script) error

// 请求/响应类型
type CreateWorkspaceRequest struct {
    Name   string                  `json:"name" binding:"required"`
    Image  string                  `json:"image"`
    Scripts []domain.Script        `json:"scripts,omitempty"`
}
```

#### ✅ 验收标准

- [ ] 可以创建工作空间，容器自动启动
- [ ] 脚本按顺序执行，失败时状态更新为 error
- [ ] 可以查询工作空间详情
- [ ] 可以列出所有工作空间
- [ ] 可以删除工作空间（容器也被删除）
- [ ] 创建失败时资源正确清理
- [ ] 通过集成测试

#### 📚 依赖

- Module 1, Module 2, Module 3a

---

### Module 4: 终端服务 (Terminal Service)

**负责人**：Agent 4
**预计时间**：3-4 天
**优先级**：🟡 中（核心功能）
**依赖**：Module 1, Module 2

#### 📦 包含组件

1. **TerminalService** (`internal/service/terminal.go`)
2. **TerminalSession** 会话管理

#### 📋 任务清单

- [ ] 实现 WebSocket 升级
- [ ] 实现 Docker Exec 创建（TTY 模式）
- [ ] 实现 Docker Exec Attach
- [ ] 实现双向数据传输（goroutine）
  - [ ] WebSocket → Docker Exec
  - [ ] Docker Exec → WebSocket
- [ ] 实现消息协议
  - [ ] `input` - 用户输入
  - [ ] `output` - 终端输出
  - [ ] `resize` - 终端大小调整
- [ ] 实现终端 resize 支持
- [ ] 实现会话管理
  - [ ] 会话创建
  - [ ] 会话清理
  - [ ] 超时处理
- [ ] 错误处理和连接关闭

#### 🔌 对外接口

```go
type TerminalService struct {
    dockerSvc *DockerService
    sessions  sync.Map // map[sessionID]*TerminalSession
}

func NewTerminalService(dockerSvc *DockerService) *TerminalService

// 会话管理
func (s *TerminalService) CreateSession(ctx context.Context, ws *websocket.Conn, containerID string) error
func (s *TerminalService) CloseSession(sessionID string) error

// 内部会话结构
type TerminalSession struct {
    ID          string
    ContainerID string
    WebSocket   *websocket.Conn
    ExecConn    types.HijackedResponse
    CreatedAt   time.Time
}

// 消息协议
type TerminalMessage struct {
    Type string `json:"type"` // "input", "output", "resize"
    Data string `json:"data,omitempty"`
    Cols int    `json:"cols,omitempty"`
    Rows int    `json:"rows,omitempty"`
}
```

#### ✅ 验收标准

- [ ] WebSocket 连接成功
- [ ] 可以执行命令并看到输出
- [ ] 支持交互式程序（vim, top）
- [ ] 终端大小调整正常
- [ ] 连接断开后资源清理
- [ ] 支持多并发会话
- [ ] 通过集成测试（使用 websocat 或自定义客户端）

#### 📚 依赖

- Module 1, Module 2
- `github.com/gorilla/websocket`
- `github.com/docker/docker/api/types`

---

### Module 5: 代理服务 (Proxy Service)

**负责人**：Agent 5
**预计时间**：2 天
**优先级**：🟢 低（可延后）
**依赖**：Module 1, Module 2

#### 📦 包含组件

1. **ProxyService** (`internal/service/proxy.go`)

#### 📋 任务清单

- [ ] 实现反向代理（使用 `httputil.ReverseProxy`）
- [ ] 获取容器 IP 地址
- [ ] 配置代理 Transport
- [ ] 实现路径重写
- [ ] 实现错误处理
- [ ] 添加日志

#### 🔌 对外接口

```go
type ProxyService struct {
    dockerSvc *DockerService
}

func NewProxyService(dockerSvc *DockerService) *ProxyService

// 代理请求
func (s *ProxyService) ProxyRequest(w http.ResponseWriter, r *http.Request, containerID string, port int) error

// 辅助方法
func (s *ProxyService) createReverseProxy(containerIP string, port int) *httputil.ReverseProxy
```

#### ✅ 验收标准

- [ ] 可以访问容器内 HTTP 服务
- [ ] 路径正确转发
- [ ] POST/PUT 等请求正常
- [ ] WebSocket 升级正常（如果需要）
- [ ] 错误情况正确处理
- [ ] 通过集成测试

#### 📚 依赖

- Module 1, Module 2
- Go 标准库 (`net/http/httputil`)

---

### Module 6: API 层 (API Layer)

**负责人**：Agent 6, Agent 7
**预计时间**：2-3 天
**优先级**：🟡 中
**依赖**：Module 1, Module 3b, Module 4, Module 5

#### 📦 包含组件

1. **Router** (`internal/api/router.go`)
2. **Workspace Handler** (`internal/api/handler/workspace.go`)
3. **Terminal Handler** (`internal/api/handler/terminal.go`)
4. **Proxy Handler** (`internal/api/handler/proxy.go`)

#### 📋 任务清单

**Agent 6 - Router + Workspace Handler**：
- [ ] 实现 Router 配置
  - [ ] 应用全局中间件
  - [ ] 配置路由分组
  - [ ] **应用 Auth 中间件到所有需要鉴权的路由**
    - [ ] `/api/*` - 所有 API 路由
    - [ ] `/ws/terminal/:id` - WebSocket 终端
    - [ ] `/forward/:id/:port/*path` - 端口转发
- [ ] 实现 Workspace Handler
  - [ ] `POST /api/workspaces` - 创建工作空间
  - [ ] `GET /api/workspaces` - 列出工作空间
  - [ ] `GET /api/workspaces/:id` - 获取工作空间
  - [ ] `DELETE /api/workspaces/:id` - 删除工作空间
- [ ] 请求验证（使用 Gin binding）
- [ ] 响应格式统一

**Agent 7 - Terminal Handler + Proxy Handler**：
- [ ] 实现 Terminal Handler
  - [ ] `GET /ws/terminal/:id` - WebSocket 终端连接
  - [ ] WebSocket 升级
  - [ ] 会话管理
- [ ] 实现 Proxy Handler
  - [ ] `ANY /forward/:id/:port/*path` - 端口转发
  - [ ] 路径解析
  - [ ] 代理请求转发

#### 🔌 对外接口

**Router**：
```go
func SetupRouter(
    cfg *config.Config,
    workspaceSvc *service.WorkspaceService,
    terminalSvc *service.TerminalService,
    proxySvc *service.ProxyService,
) *gin.Engine
```

**Handler 接口**：
```go
// Workspace Handler
type WorkspaceHandler struct {
    service *service.WorkspaceService
}

func (h *WorkspaceHandler) Create(c *gin.Context)
func (h *WorkspaceHandler) List(c *gin.Context)
func (h *WorkspaceHandler) Get(c *gin.Context)
func (h *WorkspaceHandler) Delete(c *gin.Context)

// Terminal Handler
type TerminalHandler struct {
    service *service.TerminalService
}

func (h *TerminalHandler) Connect(c *gin.Context)

// Proxy Handler
type ProxyHandler struct {
    service *service.ProxyService
}

func (h *ProxyHandler) Forward(c *gin.Context)
```

#### ✅ 验收标准

- [ ] 所有 API 正常工作
- [ ] 请求验证正确
- [ ] 错误响应格式统一
- [ ] **Auth 中间件正确应用到所有路由（/api/*, /ws/terminal/:id, /forward/:id/:port/*path）**
- [ ] **未授权请求返回 401 错误**
- [ ] WebSocket 升级成功
- [ ] 代理转发正常
- [ ] 通过 API 集成测试

#### 📚 依赖

- Module 1, Module 3b, Module 4, Module 5
- `github.com/gin-gonic/gin`

---

### Module 7: 部署和 CI/CD (Deployment)

**负责人**：Agent 8
**预计时间**：1-2 天
**优先级**：🟢 低（最后阶段）
**依赖**：所有其他模块

#### 📦 包含组件

1. **Dockerfile**
2. **docker-compose.yml**
3. **.dockerignore**
4. **GitHub Actions** (`.github/workflows/docker-build.yml`)
5. **Main 入口** (`cmd/server/main.go`)

#### 📋 任务清单

- [ ] 编写 Dockerfile（多阶段构建）
- [ ] 编写 docker-compose.yml
- [ ] 配置环境变量
- [ ] 编写 .dockerignore
- [ ] 配置 GitHub Actions
  - [ ] 自动构建 Docker 镜像
  - [ ] 推送到 ghcr.io
  - [ ] 自动打标签
- [ ] 实现 main.go 入口
  - [ ] 加载配置
  - [ ] 初始化服务
  - [ ] 启动 HTTP 服务
- [ ] 测试部署流程

#### 🔌 对外接口

**Main 入口**：
```go
func main() {
    // 1. 加载配置
    cfg := config.Load()

    // 2. 初始化服务
    dockerSvc, _ := service.NewDockerService(cfg)
    repo := repository.NewMemoryRepository()
    workspaceSvc := service.NewWorkspaceService(dockerSvc, repo, cfg)
    terminalSvc := service.NewTerminalService(dockerSvc)
    proxySvc := service.NewProxyService(dockerSvc)

    // 3. 设置路由
    router := api.SetupRouter(cfg, workspaceSvc, terminalSvc, proxySvc)

    // 4. 启动服务
    router.Run(":" + cfg.Port)
}
```

#### ✅ 验收标准

- [ ] Docker 镜像构建成功
- [ ] 容器启动正常
- [ ] 所有功能正常工作
- [ ] CI/CD 自动构建成功
- [ ] 镜像可从 ghcr.io 拉取
- [ ] docker-compose 一键部署成功

#### 📚 依赖

- 所有其他模块
- Docker
- GitHub Actions

---

## 接口定义

### API 端点详细定义

#### 1. 创建工作空间

```
POST /api/workspaces
Content-Type: application/json
Authorization: Bearer {token}
```

**请求体**：
```json
{
  "name": "my-workspace",
  "image": "ubuntu:22.04",
  "scripts": [
    {
      "name": "install-tools",
      "content": "#!/bin/bash\napt-get update && apt-get install -y curl git",
      "order": 1
    },
    {
      "name": "setup-env",
      "content": "#!/bin/bash\necho 'export PATH=/usr/local/bin:$PATH' >> ~/.bashrc",
      "order": 2
    }
  ]
}
```

**响应**：
```json
{
  "id": "ws-123abc",
  "name": "my-workspace",
  "container_id": "docker-container-id",
  "status": "creating",
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-10T12:00:00Z",
  "config": {
    "image": "ubuntu:22.04",
    "scripts": [...],
    "exposed_ports": []
  }
}
```

**错误响应**：
```json
{
  "error": "Invalid request: name is required",
  "code": "INVALID_REQUEST"
}
```

#### 2. 列出工作空间

```
GET /api/workspaces
Authorization: Bearer {token}
```

**响应**：
```json
[
  {
    "id": "ws-123abc",
    "name": "my-workspace",
    "status": "running",
    ...
  },
  {
    "id": "ws-456def",
    "name": "another-workspace",
    "status": "stopped",
    ...
  }
]
```

#### 3. 获取工作空间

```
GET /api/workspaces/:id
Authorization: Bearer {token}
```

**响应**：同创建工作空间响应

**错误响应**：
```json
{
  "error": "Workspace not found",
  "code": "NOT_FOUND"
}
```

#### 4. 删除工作空间

```
DELETE /api/workspaces/:id
Authorization: Bearer {token}
```

**响应**：
```json
{
  "message": "Workspace deleted successfully"
}
```

#### 5. WebSocket 终端连接

```
GET /ws/terminal/:id?token={token}
Upgrade: websocket
```

**消息格式**：

客户端 → 服务器：
```json
{"type": "input", "data": "ls -la\n"}
{"type": "resize", "cols": 80, "rows": 24}
```

服务器 → 客户端：
```json
{"type": "output", "data": "total 48\ndrwxr-xr-x..."}
{"type": "error", "data": "Connection lost"}
```

#### 6. 端口转发

```
GET /forward/:id/:port/*path?token={token}
POST /forward/:id/:port/*path?token={token}
...（任意 HTTP 方法）
```

直接代理到容器内服务，透明转发所有请求和响应。

---

## 开发顺序建议

### 📅 详细时间表

#### Week 1（第 1-7 天）

**Day 1-2：Round 1**
- Agent 1: 完成 Module 1（基础设施层）
- 里程碑：配置、日志、中间件可用

**Day 3-5：Round 2**
- Agent 2: 完成 Module 2（Docker 服务）
- Agent 3: 完成 Module 3a（数据层）
- 里程碑：Docker 操作正常，数据存储可用

**Day 6-7：Round 3 开始**
- Agent 3: 开始 Module 3b（工作空间服务）
- Agent 4: 开始 Module 4（终端服务）
- Agent 5: 开始 Module 5（代理服务）

#### Week 2（第 8-14 天）

**Day 8-11：Round 3 继续**
- Agent 3, 4, 5: 继续并行开发业务服务
- 里程碑：三大核心服务完成

**Day 12-13：Round 4**
- Agent 6: 完成 Router + Workspace Handler
- Agent 7: 完成 Terminal Handler + Proxy Handler
- 里程碑：API 层完成，系统集成

**Day 14：Round 5**
- Agent 8: 完成部署和 CI/CD
- 里程碑：系统可部署

#### Week 3（第 15-16 天，缓冲）

**Day 15-16：集成测试和优化**
- 全员：集成测试、bug 修复、性能优化
- 里程碑：第一阶段完成

---

## 团队协作建议

### 📢 沟通机制

1. **接口先行**：每个模块先定义清晰的接口
2. **Mock 测试**：依赖未完成时使用 Mock
3. **每日同步**：每天同步进度和阻塞点
4. **集成测试**：Round 结束时进行集成测试

### 🔧 工具建议

- **代码仓库**：Git + GitHub
- **分支策略**：
  - `main` - 主分支
  - `module-1-foundation` - Module 1 开发分支
  - `module-2-docker` - Module 2 开发分支
  - ...（每个模块一个分支）
- **Pull Request**：每个模块完成后提交 PR
- **Code Review**：至少一人 review 后合并

### ✅ 质量保证

- **单元测试**：每个模块至少 70% 覆盖率
- **集成测试**：Round 结束时进行
- **代码规范**：使用 `gofmt`, `golangci-lint`
- **文档**：每个模块提供 README 和示例

---

## 风险管理

### ⚠️ 潜在风险

1. **模块依赖阻塞**
   - 缓解：严格按 Round 顺序，依赖未完成时使用 Mock

2. **接口不匹配**
   - 缓解：Round 1 完成后统一 review 所有接口定义

3. **集成问题**
   - 缓解：每个 Round 结束进行集成测试

4. **时间延期**
   - 缓解：设置 Week 3 作为缓冲时间

### 🎯 成功关键

1. **接口定义清晰**：避免后期返工
2. **严格依赖管理**：按顺序开发
3. **充分测试**：单元测试 + 集成测试
4. **持续集成**：每个模块完成后立即集成

---

## 总结

通过将后端开发拆分为 7 个独立模块，我们可以实现：

- ✅ **并行开发**：最多 5 个 agent 同时工作（Round 3）
- ✅ **降低风险**：模块独立，依赖清晰
- ✅ **提高质量**：每个模块独立测试
- ✅ **加速交付**：预计 10-16 天完成（vs 原计划 15-21 天）

**下一步行动**：
1. 组建开发团队（分配 Agent）
2. Review 接口定义
3. 开始 Round 1 开发

---

**文档版本**：v1.0
**创建日期**：2025-11-10
**维护者**：ViBox Team
