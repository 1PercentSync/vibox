# ViBox Go 后端架构设计

## 项目结构

```
vibox/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── api/
│   │   ├── handler/
│   │   │   ├── workspace.go     # 工作空间 API Handler
│   │   │   ├── terminal.go      # WebSocket 终端 Handler
│   │   │   └── proxy.go         # 端口转发 Handler
│   │   ├── middleware/
│   │   │   ├── cors.go          # CORS 中间件
│   │   │   ├── logger.go        # 日志中间件
│   │   │   └── recovery.go      # 错误恢复中间件
│   │   └── router.go            # 路由配置
│   ├── domain/
│   │   ├── workspace.go         # 工作空间领域模型
│   │   └── script.go            # 脚本领域模型
│   ├── service/
│   │   ├── workspace.go         # 工作空间业务逻辑
│   │   ├── docker.go            # Docker 操作封装
│   │   ├── terminal.go          # 终端会话管理
│   │   └── proxy.go             # 代理服务
│   ├── repository/
│   │   └── workspace.go         # 数据持久化（内存/文件/数据库）
│   └── config/
│       └── config.go            # 配置管理
├── pkg/
│   └── utils/
│       ├── id.go                # ID 生成工具
│       └── logger.go            # 日志工具
├── web/                         # 前端静态文件（待定）
├── scripts/                     # 部署脚本
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

---

## 技术栈

### 核心依赖

```go
require (
    github.com/gin-gonic/gin v1.9.1           // Web 框架
    github.com/docker/docker v24.0.0          // Docker SDK
    github.com/gorilla/websocket v1.5.1       // WebSocket
    github.com/google/uuid v1.5.0             // UUID 生成
)
```

### 标准库使用

```go
import (
    "net/http/httputil"  // 反向代理
    "context"            // 上下文管理
    "io"                 // 流式 I/O
)
```

---

## 核心模块设计

### 1. 领域模型（Domain）

#### workspace.go

```go
package domain

import "time"

// WorkspaceStatus 工作空间状态
type WorkspaceStatus string

const (
    StatusCreating WorkspaceStatus = "creating"
    StatusRunning  WorkspaceStatus = "running"
    StatusStopped  WorkspaceStatus = "stopped"
    StatusError    WorkspaceStatus = "error"
)

// Workspace 工作空间领域模型
type Workspace struct {
    ID          string          `json:"id"`
    Name        string          `json:"name"`
    ContainerID string          `json:"container_id"`
    Status      WorkspaceStatus `json:"status"`
    CreatedAt   time.Time       `json:"created_at"`
    Config      WorkspaceConfig `json:"config"`
}

// WorkspaceConfig 工作空间配置
type WorkspaceConfig struct {
    Image        string         `json:"image"`         // 基础镜像，默认 ubuntu:22.04
    Scripts      []Script       `json:"scripts"`       // 初始化脚本
    ExposedPorts []ExposedPort  `json:"exposed_ports"` // 暴露的端口
}

// Script 脚本
type Script struct {
    Name    string `json:"name"`
    Content string `json:"content"` // Base64 编码的脚本内容
    Order   int    `json:"order"`   // 执行顺序
}

// ExposedPort 暴露的端口
type ExposedPort struct {
    ContainerPort int    `json:"container_port"`
    Enabled       bool   `json:"enabled"`
    PublicPath    string `json:"public_path"` // 如 /forward/workspace-id/8080
}

// CreateWorkspaceRequest 创建工作空间请求
type CreateWorkspaceRequest struct {
    Name    string         `json:"name" binding:"required"`
    Image   string         `json:"image"`
    Scripts []Script       `json:"scripts"`
}

// CreateWorkspaceResponse 创建工作空间响应
type CreateWorkspaceResponse struct {
    Workspace *Workspace `json:"workspace"`
}
```

---

### 2. 服务层（Service）

#### docker.go - Docker 操作封装

```go
package service

import (
    "context"
    "io"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)

// DockerService Docker 操作服务
type DockerService struct {
    client *client.Client
}

// NewDockerService 创建 Docker 服务
func NewDockerService() (*DockerService, error) {
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        return nil, err
    }
    return &DockerService{client: cli}, nil
}

// CreateContainer 创建容器
func (s *DockerService) CreateContainer(ctx context.Context, config *ContainerConfig) (string, error) {
    // 拉取镜像（如果不存在）
    reader, err := s.client.ImagePull(ctx, config.Image, types.ImagePullOptions{})
    if err != nil {
        return "", err
    }
    defer reader.Close()
    io.Copy(io.Discard, reader) // 等待拉取完成

    // 创建容器配置
    containerConfig := &container.Config{
        Image: config.Image,
        Cmd:   []string{"/bin/bash", "-c", "tail -f /dev/null"}, // 保持容器运行
        Tty:   true,
        Labels: map[string]string{
            "vibox.workspace": config.WorkspaceID,
        },
    }

    hostConfig := &container.HostConfig{
        // 不映射端口到宿主机，通过后端反向代理访问
        NetworkMode: "bridge",
        // 资源限制
        Resources: container.Resources{
            Memory:   config.MemoryLimit,   // 如 512MB
            NanoCPUs: config.CPULimit,      // 如 1 核
        },
    }

    resp, err := s.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, config.Name)
    if err != nil {
        return "", err
    }

    return resp.ID, nil
}

// StartContainer 启动容器
func (s *DockerService) StartContainer(ctx context.Context, containerID string) error {
    return s.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

// StopContainer 停止容器
func (s *DockerService) StopContainer(ctx context.Context, containerID string) error {
    timeout := 10 // 秒
    return s.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

// RemoveContainer 删除容器
func (s *DockerService) RemoveContainer(ctx context.Context, containerID string) error {
    return s.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
        Force: true,
    })
}

// ExecInContainer 在容器中执行命令
func (s *DockerService) ExecInContainer(ctx context.Context, containerID string, cmd []string) error {
    exec, err := s.client.ContainerExecCreate(ctx, containerID, types.ExecConfig{
        Cmd:          cmd,
        AttachStdout: true,
        AttachStderr: true,
    })
    if err != nil {
        return err
    }

    return s.client.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{})
}

// CopyToContainer 复制文件到容器
func (s *DockerService) CopyToContainer(ctx context.Context, containerID, path string, content io.Reader) error {
    return s.client.CopyToContainer(ctx, containerID, path, content, types.CopyToContainerOptions{})
}

// GetContainerIP 获取容器 IP
func (s *DockerService) GetContainerIP(ctx context.Context, containerID string) (string, error) {
    inspect, err := s.client.ContainerInspect(ctx, containerID)
    if err != nil {
        return "", err
    }
    return inspect.NetworkSettings.IPAddress, nil
}

// ContainerConfig 容器配置
type ContainerConfig struct {
    WorkspaceID string
    Name        string
    Image       string
    MemoryLimit int64
    CPULimit    int64
}
```

#### workspace.go - 工作空间业务逻辑

```go
package service

import (
    "context"
    "fmt"

    "vibox/internal/domain"
    "vibox/internal/repository"
)

// WorkspaceService 工作空间服务
type WorkspaceService struct {
    docker *DockerService
    repo   repository.WorkspaceRepository
}

// NewWorkspaceService 创建工作空间服务
func NewWorkspaceService(docker *DockerService, repo repository.WorkspaceRepository) *WorkspaceService {
    return &WorkspaceService{
        docker: docker,
        repo:   repo,
    }
}

// CreateWorkspace 创建工作空间
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req *domain.CreateWorkspaceRequest) (*domain.Workspace, error) {
    // 生成工作空间 ID
    workspaceID := generateID()

    // 创建工作空间对象
    workspace := &domain.Workspace{
        ID:     workspaceID,
        Name:   req.Name,
        Status: domain.StatusCreating,
        Config: domain.WorkspaceConfig{
            Image:   req.Image,
            Scripts: req.Scripts,
        },
    }

    // 保存到存储
    if err := s.repo.Save(workspace); err != nil {
        return nil, err
    }

    // 异步创建容器
    go s.createContainerAsync(workspace)

    return workspace, nil
}

// createContainerAsync 异步创建容器
func (s *WorkspaceService) createContainerAsync(workspace *domain.Workspace) {
    ctx := context.Background()

    // 创建容器
    containerID, err := s.docker.CreateContainer(ctx, &ContainerConfig{
        WorkspaceID: workspace.ID,
        Name:        fmt.Sprintf("vibox-%s", workspace.ID),
        Image:       workspace.Config.Image,
        MemoryLimit: 512 * 1024 * 1024,  // 512MB
        CPULimit:    1000000000,          // 1 核
    })
    if err != nil {
        workspace.Status = domain.StatusError
        s.repo.Save(workspace)
        return
    }

    workspace.ContainerID = containerID

    // 启动容器
    if err := s.docker.StartContainer(ctx, containerID); err != nil {
        workspace.Status = domain.StatusError
        s.repo.Save(workspace)
        return
    }

    // 执行初始化脚本
    if err := s.executeScripts(ctx, workspace); err != nil {
        workspace.Status = domain.StatusError
        s.repo.Save(workspace)
        return
    }

    workspace.Status = domain.StatusRunning
    s.repo.Save(workspace)
}

// executeScripts 执行初始化脚本
func (s *WorkspaceService) executeScripts(ctx context.Context, workspace *domain.Workspace) error {
    // 按顺序执行脚本
    for _, script := range workspace.Config.Scripts {
        // 将脚本复制到容器
        // 执行脚本
        // ...
    }
    return nil
}

// GetWorkspace 获取工作空间
func (s *WorkspaceService) GetWorkspace(id string) (*domain.Workspace, error) {
    return s.repo.Get(id)
}

// ListWorkspaces 列出所有工作空间
func (s *WorkspaceService) ListWorkspaces() ([]*domain.Workspace, error) {
    return s.repo.List()
}

// DeleteWorkspace 删除工作空间
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {
    workspace, err := s.repo.Get(id)
    if err != nil {
        return err
    }

    // 删除容器
    if workspace.ContainerID != "" {
        if err := s.docker.RemoveContainer(ctx, workspace.ContainerID); err != nil {
            return err
        }
    }

    // 从存储删除
    return s.repo.Delete(id)
}
```

#### terminal.go - 终端会话管理

```go
package service

import (
    "context"
    "io"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/client"
    "github.com/gorilla/websocket"
)

// TerminalService 终端服务
type TerminalService struct {
    docker *DockerService
}

// NewTerminalService 创建终端服务
func NewTerminalService(docker *DockerService) *TerminalService {
    return &TerminalService{docker: docker}
}

// TerminalSession 终端会话
type TerminalSession struct {
    ws          *websocket.Conn
    containerID string
    execID      string
    hijacked    types.HijackedResponse
}

// CreateSession 创建终端会话
func (s *TerminalService) CreateSession(ws *websocket.Conn, containerID string) (*TerminalSession, error) {
    ctx := context.Background()

    // 创建 exec 实例
    exec, err := s.docker.client.ContainerExecCreate(ctx, containerID, types.ExecConfig{
        Cmd:          []string{"/bin/bash"},
        AttachStdin:  true,
        AttachStdout: true,
        AttachStderr: true,
        Tty:          true,
    })
    if err != nil {
        return nil, err
    }

    // 连接到 exec
    hijacked, err := s.docker.client.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{
        Tty: true,
    })
    if err != nil {
        return nil, err
    }

    session := &TerminalSession{
        ws:          ws,
        containerID: containerID,
        execID:      exec.ID,
        hijacked:    hijacked,
    }

    // 启动双向数据传输
    go session.handleInput()
    go session.handleOutput()

    return session, nil
}

// handleInput WebSocket → 容器
func (session *TerminalSession) handleInput() {
    defer session.Close()

    for {
        _, message, err := session.ws.ReadMessage()
        if err != nil {
            return
        }

        // 解析消息类型
        var msg TerminalMessage
        if err := json.Unmarshal(message, &msg); err != nil {
            continue
        }

        switch msg.Type {
        case "input":
            // 发送到容器
            session.hijacked.Conn.Write([]byte(msg.Data))
        case "resize":
            // 调整终端大小
            // ...
        }
    }
}

// handleOutput 容器 → WebSocket
func (session *TerminalSession) handleOutput() {
    defer session.Close()

    buf := make([]byte, 4096)
    for {
        n, err := session.hijacked.Reader.Read(buf)
        if err != nil {
            if err != io.EOF {
                // 记录错误
            }
            return
        }

        // 发送到 WebSocket
        msg := TerminalMessage{
            Type: "output",
            Data: string(buf[:n]),
        }
        if err := session.ws.WriteJSON(msg); err != nil {
            return
        }
    }
}

// Close 关闭会话
func (session *TerminalSession) Close() {
    session.hijacked.Close()
    session.ws.Close()
}

// TerminalMessage WebSocket 消息格式
type TerminalMessage struct {
    Type string `json:"type"` // input, output, resize
    Data string `json:"data"`
    Cols int    `json:"cols,omitempty"`
    Rows int    `json:"rows,omitempty"`
}
```

#### proxy.go - 端口转发服务

```go
package service

import (
    "fmt"
    "net/http"
    "net/http/httputil"
    "net/url"
)

// ProxyService 代理服务
type ProxyService struct {
    docker *DockerService
}

// NewProxyService 创建代理服务
func NewProxyService(docker *DockerService) *ProxyService {
    return &ProxyService{docker: docker}
}

// ProxyRequest 代理请求到容器
func (s *ProxyService) ProxyRequest(w http.ResponseWriter, r *http.Request, containerID string, port int) error {
    // 获取容器 IP
    containerIP, err := s.docker.GetContainerIP(r.Context(), containerID)
    if err != nil {
        return err
    }

    // 构建目标 URL
    targetURL, err := url.Parse(fmt.Sprintf("http://%s:%d", containerIP, port))
    if err != nil {
        return err
    }

    // 创建反向代理
    proxy := httputil.NewSingleHostReverseProxy(targetURL)

    // 自定义错误处理
    proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
        http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
    }

    // 代理请求
    proxy.ServeHTTP(w, r)
    return nil
}
```

---

### 3. API Handler

#### workspace.go

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "vibox/internal/domain"
    "vibox/internal/service"
)

// WorkspaceHandler 工作空间处理器
type WorkspaceHandler struct {
    service *service.WorkspaceService
}

// NewWorkspaceHandler 创建处理器
func NewWorkspaceHandler(service *service.WorkspaceService) *WorkspaceHandler {
    return &WorkspaceHandler{service: service}
}

// Create 创建工作空间
func (h *WorkspaceHandler) Create(c *gin.Context) {
    var req domain.CreateWorkspaceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 设置默认镜像
    if req.Image == "" {
        req.Image = "ubuntu:22.04"
    }

    workspace, err := h.service.CreateWorkspace(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, workspace)
}

// Get 获取工作空间
func (h *WorkspaceHandler) Get(c *gin.Context) {
    id := c.Param("id")

    workspace, err := h.service.GetWorkspace(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
        return
    }

    c.JSON(http.StatusOK, workspace)
}

// List 列出工作空间
func (h *WorkspaceHandler) List(c *gin.Context) {
    workspaces, err := h.service.ListWorkspaces()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, workspaces)
}

// Delete 删除工作空间
func (h *WorkspaceHandler) Delete(c *gin.Context) {
    id := c.Param("id")

    if err := h.service.DeleteWorkspace(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "workspace deleted"})
}
```

#### terminal.go

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "vibox/internal/service"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // 生产环境需要严格检查
    },
}

// TerminalHandler 终端处理器
type TerminalHandler struct {
    terminalService  *service.TerminalService
    workspaceService *service.WorkspaceService
}

// NewTerminalHandler 创建处理器
func NewTerminalHandler(ts *service.TerminalService, ws *service.WorkspaceService) *TerminalHandler {
    return &TerminalHandler{
        terminalService:  ts,
        workspaceService: ws,
    }
}

// Connect 连接到终端
func (h *TerminalHandler) Connect(c *gin.Context) {
    workspaceID := c.Param("id")

    // 获取工作空间
    workspace, err := h.workspaceService.GetWorkspace(workspaceID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
        return
    }

    // 升级为 WebSocket
    ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }

    // 创建终端会话
    _, err = h.terminalService.CreateSession(ws, workspace.ContainerID)
    if err != nil {
        ws.Close()
        return
    }

    // 会话会在 goroutine 中自动运行
}
```

#### proxy.go

```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "vibox/internal/service"
)

// ProxyHandler 代理处理器
type ProxyHandler struct {
    proxyService     *service.ProxyService
    workspaceService *service.WorkspaceService
}

// NewProxyHandler 创建处理器
func NewProxyHandler(ps *service.ProxyService, ws *service.WorkspaceService) *ProxyHandler {
    return &ProxyHandler{
        proxyService:     ps,
        workspaceService: ws,
    }
}

// Forward 转发请求到容器
func (h *ProxyHandler) Forward(c *gin.Context) {
    workspaceID := c.Param("id")
    portStr := c.Param("port")

    port, err := strconv.Atoi(portStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid port"})
        return
    }

    // 获取工作空间
    workspace, err := h.workspaceService.GetWorkspace(workspaceID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
        return
    }

    // 代理请求
    if err := h.proxyService.ProxyRequest(c.Writer, c.Request, workspace.ContainerID, port); err != nil {
        c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
    }
}
```

---

### 4. 路由配置

#### router.go

```go
package api

import (
    "github.com/gin-gonic/gin"
    "vibox/internal/api/handler"
    "vibox/internal/api/middleware"
)

// SetupRouter 配置路由
func SetupRouter(
    workspaceHandler *handler.WorkspaceHandler,
    terminalHandler *handler.TerminalHandler,
    proxyHandler *handler.ProxyHandler,
) *gin.Engine {
    r := gin.New()

    // 中间件
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())

    // API 路由
    api := r.Group("/api")
    {
        // 工作空间
        workspaces := api.Group("/workspaces")
        {
            workspaces.POST("", workspaceHandler.Create)
            workspaces.GET("", workspaceHandler.List)
            workspaces.GET("/:id", workspaceHandler.Get)
            workspaces.DELETE("/:id", workspaceHandler.Delete)
        }
    }

    // WebSocket 终端
    r.GET("/ws/terminal/:id", terminalHandler.Connect)

    // 端口转发
    r.Any("/forward/:id/:port/*path", proxyHandler.Forward)

    // 前端静态文件（待实现）
    // r.Static("/", "./web")

    return r
}
```

---

### 5. 程序入口

#### main.go

```go
package main

import (
    "log"

    "vibox/internal/api"
    "vibox/internal/api/handler"
    "vibox/internal/repository"
    "vibox/internal/service"
)

func main() {
    // 初始化 Docker 服务
    dockerService, err := service.NewDockerService()
    if err != nil {
        log.Fatal(err)
    }

    // 初始化存储
    workspaceRepo := repository.NewInMemoryWorkspaceRepository()

    // 初始化服务
    workspaceService := service.NewWorkspaceService(dockerService, workspaceRepo)
    terminalService := service.NewTerminalService(dockerService)
    proxyService := service.NewProxyService(dockerService)

    // 初始化处理器
    workspaceHandler := handler.NewWorkspaceHandler(workspaceService)
    terminalHandler := handler.NewTerminalHandler(terminalService, workspaceService)
    proxyHandler := handler.NewProxyHandler(proxyService, workspaceService)

    // 配置路由
    router := api.SetupRouter(workspaceHandler, terminalHandler, proxyHandler)

    // 启动服务
    log.Println("ViBox server starting on :3000")
    if err := router.Run(":3000"); err != nil {
        log.Fatal(err)
    }
}
```

---

## 数据持久化

MVP 阶段使用内存存储，后续可扩展为文件或数据库。

```go
package repository

import (
    "errors"
    "sync"

    "vibox/internal/domain"
)

// WorkspaceRepository 工作空间存储接口
type WorkspaceRepository interface {
    Save(workspace *domain.Workspace) error
    Get(id string) (*domain.Workspace, error)
    List() ([]*domain.Workspace, error)
    Delete(id string) error
}

// InMemoryWorkspaceRepository 内存存储实现
type InMemoryWorkspaceRepository struct {
    mu    sync.RWMutex
    store map[string]*domain.Workspace
}

// NewInMemoryWorkspaceRepository 创建内存存储
func NewInMemoryWorkspaceRepository() *InMemoryWorkspaceRepository {
    return &InMemoryWorkspaceRepository{
        store: make(map[string]*domain.Workspace),
    }
}

func (r *InMemoryWorkspaceRepository) Save(workspace *domain.Workspace) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.store[workspace.ID] = workspace
    return nil
}

func (r *InMemoryWorkspaceRepository) Get(id string) (*domain.Workspace, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    ws, ok := r.store[id]
    if !ok {
        return nil, errors.New("workspace not found")
    }
    return ws, nil
}

func (r *InMemoryWorkspaceRepository) List() ([]*domain.Workspace, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    list := make([]*domain.Workspace, 0, len(r.store))
    for _, ws := range r.store {
        list = append(list, ws)
    }
    return list, nil
}

func (r *InMemoryWorkspaceRepository) Delete(id string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.store, id)
    return nil
}
```

---

## 配置管理

```go
package config

import "os"

// Config 应用配置
type Config struct {
    Port         string
    DockerHost   string
    DefaultImage string
    MemoryLimit  int64
    CPULimit     int64
}

// Load 加载配置
func Load() *Config {
    return &Config{
        Port:         getEnv("PORT", "3000"),
        DockerHost:   getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
        DefaultImage: getEnv("DEFAULT_IMAGE", "ubuntu:22.04"),
        MemoryLimit:  512 * 1024 * 1024, // 512MB
        CPULimit:     1000000000,         // 1 CPU
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

---

## 总结

这个架构设计：

1. **清晰的分层**：Domain → Service → Handler
2. **依赖注入**：便于测试和扩展
3. **接口抽象**：Repository 可以轻松切换实现
4. **并发安全**：使用 goroutine 和锁保护
5. **标准库优先**：反向代理用标准库，减少依赖

相关文档：
- [WebSSH 实现原理](./WEBSSH_PRINCIPLE.md)
- [开发计划](./DEVELOPMENT_PLAN.md)
