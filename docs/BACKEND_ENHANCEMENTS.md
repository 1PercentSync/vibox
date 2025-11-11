# Backend Enhancements - 后端增强方案

> **目标**：为 ViBox 后端添加端口标签、容器重置和数据持久化功能
>
> **版本**：v1.1.0
>
> **日期**：2025-11-10

---

## 目录

1. [需求概述](#需求概述)
2. [架构变更](#架构变更)
3. [数据模型变更](#数据模型变更)
4. [API 变更](#api-变更)
5. [实现细节](#实现细节)
6. [迁移方案](#迁移方案)
7. [测试计划](#测试计划)

---

## 需求概述

### 1. 端口标签功能

**背景**：
虽然后端采用动态端口访问（无需预先声明），但前端需要为常用端口提供快捷访问按钮。

**需求**：
- ✅ 创建工作空间时可以设定 `端口:服务名` 映射（例如 `8080: "VS Code Server"`）
- ✅ 支持后续更新端口映射列表
- ✅ 前端根据此列表显示快捷按钮
- ✅ 用户仍可通过修改 URL 手动访问任意端口

**示例**：
```json
{
  "name": "dev-env",
  "image": "ubuntu:22.04",
  "ports": {
    "8080": "VS Code Server",
    "3000": "Web App",
    "5432": "PostgreSQL"
  }
}
```

前端显示：
- [VS Code Server] → `/forward/ws-xxx/8080/`
- [Web App] → `/forward/ws-xxx/3000/`
- [PostgreSQL] → `/forward/ws-xxx/5432/`

### 2. 容器重置功能

**背景**：
用户可能需要将工作空间恢复到初始状态，重新执行初始化脚本。

**需求**：
- ✅ 提供 API 端点重置容器
- ✅ 删除旧容器，按照创建时的配置重新创建
- ✅ 重新执行所有初始化脚本
- ✅ 保留工作空间 ID 和配置

**使用场景**：
- 脚本执行失败，想重新运行
- 容器状态混乱，想恢复干净环境
- 测试脚本，需要多次重置

### 3. 数据持久化功能

**背景**：
当前工作空间数据存储在内存中，主容器重启后丢失。

**需求**：
- ✅ 持久化工作空间配置到磁盘（JSON 文件）
- ✅ 主容器重启时自动加载配置并重新创建所有工作空间
- ✅ 主容器退出时自动删除所有工作空间容器
- ✅ 启动时清理异常残留的旧容器

**持久化内容**：
- 工作空间 ID、名称、配置
- 端口映射
- 创建时间
- 脚本内容

**不持久化内容**：
- 容器 ID（主容器退出时删除，重启时重建）
- 容器状态（重启后重新创建）
- 更新时间（不需要）

---

## 架构变更

### 变更前（v1.0.0）

```
┌─────────────────────────────────────┐
│  WorkspaceService                   │
│  ├── CreateWorkspace()              │
│  ├── GetWorkspace()                 │
│  ├── ListWorkspaces()               │
│  └── DeleteWorkspace()              │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  Repository (内存存储)               │
│  map[string]*Workspace              │
└─────────────────────────────────────┘
```

**问题**：
- 数据存储在内存中，重启丢失
- 无法恢复工作空间

### 变更后（v1.1.0）

```
┌─────────────────────────────────────┐
│  WorkspaceService                   │
│  ├── CreateWorkspace()              │
│  ├── GetWorkspace()                 │
│  ├── ListWorkspaces()               │
│  ├── DeleteWorkspace()              │
│  ├── UpdatePorts()          ← 新增  │
│  ├── ResetWorkspace()       ← 新增  │
│  ├── RestoreWorkspaces()    ← 新增  │
│  ├── CleanupContainers()    ← 新增  │
│  └── Shutdown()             ← 新增  │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  Repository (文件存储)               │
│  ├── Save()                 ← 新增  │
│  ├── Load()                 ← 新增  │
│  └── data/workspaces.json   ← 新增  │
└─────────────────────────────────────┘
```

**优势**：
- ✅ 数据持久化到磁盘
- ✅ 重启后自动恢复
- ✅ 支持容器重置

---

## 数据模型变更

### 1. Workspace 结构体

#### 变更前

```go
// internal/domain/workspace.go
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
    Image   string   `json:"image"`
    Scripts []Script `json:"scripts"`
}
```

#### 变更后

```go
// internal/domain/workspace.go
type Workspace struct {
    ID          string                `json:"id"`
    Name        string                `json:"name"`
    ContainerID string                `json:"container_id,omitempty"` // 运行时字段，不持久化
    Status      WorkspaceStatus       `json:"status,omitempty"`       // 运行时字段，不持久化
    CreatedAt   time.Time             `json:"created_at"`
    Config      WorkspaceConfig       `json:"config"`
    Ports       map[string]string     `json:"ports,omitempty"`        // 新增：端口映射
    Error       string                `json:"error,omitempty"`        // 运行时字段，不持久化
}

type WorkspaceConfig struct {
    Image   string   `json:"image"`
    Scripts []Script `json:"scripts"`
}

type Script struct {
    Name    string `json:"name"`
    Content string `json:"content"`
    Order   int    `json:"order"`
}
```

**字段说明**：

| 字段 | 类型 | 持久化 | 说明 |
|------|------|--------|------|
| `Ports` | `map[string]string` | ✅ | 端口映射，key=端口号，value=服务名 |
| `ContainerID` | `string` | ❌ | 运行时字段，主容器退出时删除容器 |
| `Status` | `WorkspaceStatus` | ❌ | 运行时字段，重启后重新检测 |
| `Error` | `string` | ❌ | 运行时错误信息 |

### 2. 持久化数据结构

```go
// internal/repository/workspace.go
type PersistentData struct {
    Workspaces map[string]*Workspace `json:"workspaces"` // 工作空间列表
}
```

**存储位置**：
- 开发环境：`./data/workspaces.json`
- 生产环境（Docker）：`/data/workspaces.json`（挂载卷）

**存储内容示例**：
```json
{
  "workspaces": {
    "ws-a1b2c3d4": {
      "id": "ws-a1b2c3d4",
      "name": "dev-env",
      "created_at": "2025-11-10T12:00:00Z",
      "config": {
        "image": "ubuntu:22.04",
        "scripts": [...]
      },
      "ports": {
        "8080": "VS Code Server"
      }
    }
  }
}
```

**注意**：`container_id`、`status`、`error` 等运行时字段不会被持久化。

---

## API 变更

### 1. 创建工作空间（增强）

#### 变更前

```http
POST /api/workspaces
Content-Type: application/json

{
  "name": "dev-env",
  "image": "ubuntu:22.04",
  "scripts": [...]
}
```

#### 变更后

```http
POST /api/workspaces
Content-Type: application/json

{
  "name": "dev-env",
  "image": "ubuntu:22.04",
  "scripts": [...],
  "ports": {                    // 新增：可选
    "8080": "VS Code Server",
    "3000": "Web App"
  }
}
```

**响应**：
```json
{
  "id": "ws-a1b2c3d4",
  "name": "dev-env",
  "container_id": "docker-abc123",
  "status": "creating",
  "created_at": "2025-11-10T12:00:00Z",
  "config": {
    "image": "ubuntu:22.04",
    "scripts": [...]
  },
  "ports": {
    "8080": "VS Code Server",
    "3000": "Web App"
  }
}
```

---

### 2. 更新端口映射（新增）

```http
PUT /api/workspaces/:id/ports
X-ViBox-Token: {token}
Content-Type: application/json

{
  "ports": {
    "8080": "VS Code Server",
    "3000": "Web App",
    "5432": "PostgreSQL"
  }
}
```

**响应**：
```json
{
  "id": "ws-a1b2c3d4",
  "ports": {
    "8080": "VS Code Server",
    "3000": "Web App",
    "5432": "PostgreSQL"
  }
}
```

**错误响应**：
```http
HTTP/1.1 404 Not Found

{
  "error": "Workspace not found",
  "code": "NOT_FOUND"
}
```

---

### 3. 重置工作空间（新增）

```http
POST /api/workspaces/:id/reset
X-ViBox-Token: {token}
```

**功能**：
1. 停止并删除旧容器
2. 使用原始配置创建新容器
3. 重新执行所有初始化脚本
4. 保留工作空间 ID 和配置

**响应**：
```json
{
  "id": "ws-a1b2c3d4",
  "name": "dev-env",
  "container_id": "docker-new123",
  "status": "creating",
  "message": "Workspace reset successfully"
}
```

**错误响应**：
```http
HTTP/1.1 404 Not Found

{
  "error": "Workspace not found",
  "code": "NOT_FOUND"
}
```

---

### 4. 获取工作空间（增强）

#### 变更前

```json
{
  "id": "ws-a1b2c3d4",
  "name": "dev-env",
  "status": "running",
  ...
}
```

#### 变更后

```json
{
  "id": "ws-a1b2c3d4",
  "name": "dev-env",
  "status": "running",
  "ports": {                    // 新增
    "8080": "VS Code Server"
  },
  "auto_restore": true,         // 新增
  ...
}
```

---

## 实现细节

### 1. Repository 层：持久化存储

```go
// internal/repository/workspace.go
package repository

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sync"

    "github.com/1PercentSync/vibox/internal/domain"
)

type WorkspaceRepository struct {
    workspaces map[string]*domain.Workspace
    mu         sync.RWMutex
    dataFile   string
}

func NewWorkspaceRepository(dataDir string) *WorkspaceRepository {
    if err := os.MkdirAll(dataDir, 0755); err != nil {
        panic(err)
    }

    repo := &WorkspaceRepository{
        workspaces: make(map[string]*domain.Workspace),
        dataFile:   filepath.Join(dataDir, "workspaces.json"),
    }

    // 启动时加载数据
    if err := repo.Load(); err != nil {
        // 文件不存在或损坏，使用空数据
        repo.workspaces = make(map[string]*domain.Workspace)
    }

    return repo
}

// Save saves all workspaces to disk
func (r *WorkspaceRepository) Save() error {
    r.mu.RLock()
    defer r.mu.RUnlock()

    data := PersistentData{
        Workspaces: r.workspaces,
    }

    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return err
    }

    // 原子写入：先写临时文件，再重命名
    tmpFile := r.dataFile + ".tmp"
    if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
        return err
    }

    return os.Rename(tmpFile, r.dataFile)
}

// Load loads workspaces from disk
func (r *WorkspaceRepository) Load() error {
    r.mu.Lock()
    defer r.mu.Unlock()

    jsonData, err := os.ReadFile(r.dataFile)
    if err != nil {
        return err
    }

    var data PersistentData
    if err := json.Unmarshal(jsonData, &data); err != nil {
        return err
    }

    r.workspaces = data.Workspaces
    return nil
}

// Create creates a new workspace and saves to disk
func (r *WorkspaceRepository) Create(ws *domain.Workspace) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.workspaces[ws.ID] = ws
    return r.Save()
}

// Update updates a workspace and saves to disk
func (r *WorkspaceRepository) Update(ws *domain.Workspace) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, exists := r.workspaces[ws.ID]; !exists {
        return ErrNotFound
    }

    ws.UpdatedAt = time.Now()
    r.workspaces[ws.ID] = ws
    return r.Save()
}

// Delete deletes a workspace and saves to disk
func (r *WorkspaceRepository) Delete(id string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, exists := r.workspaces[id]; !exists {
        return ErrNotFound
    }

    delete(r.workspaces, id)
    return r.Save()
}

// Get retrieves a workspace by ID
func (r *WorkspaceRepository) Get(id string) (*domain.Workspace, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    ws, exists := r.workspaces[id]
    if !exists {
        return nil, ErrNotFound
    }

    return ws, nil
}

// List retrieves all workspaces
func (r *WorkspaceRepository) List() ([]*domain.Workspace, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    result := make([]*domain.Workspace, 0, len(r.workspaces))
    for _, ws := range r.workspaces {
        result = append(result, ws)
    }

    return result, nil
}
```

---

### 2. Service 层：新增功能

```go
// internal/service/workspace.go

// UpdatePorts updates the port mappings for a workspace
func (s *WorkspaceService) UpdatePorts(ctx context.Context, id string, ports map[string]string) error {
    ws, err := s.repo.Get(id)
    if err != nil {
        return err
    }

    ws.Ports = ports
    ws.UpdatedAt = time.Now()

    return s.repo.Update(ws)
}

// ResetWorkspace resets a workspace to initial state
func (s *WorkspaceService) ResetWorkspace(ctx context.Context, id string) error {
    ws, err := s.repo.Get(id)
    if err != nil {
        return err
    }

    // 1. 删除旧容器（如果存在）
    if ws.ContainerID != "" {
        if err := s.docker.StopContainer(ctx, ws.ContainerID); err != nil {
            // 容器可能已停止，忽略错误
        }
        if err := s.docker.RemoveContainer(ctx, ws.ContainerID); err != nil {
            // 容器可能已删除，忽略错误
        }
    }

    // 2. 重置状态
    ws.ContainerID = ""
    ws.Status = domain.WorkspaceStatusCreating
    ws.Error = ""
    ws.UpdatedAt = time.Now()

    if err := s.repo.Update(ws); err != nil {
        return err
    }

    // 3. 重新创建容器
    containerID, err := s.docker.CreateContainer(ctx, ws.Config)
    if err != nil {
        ws.Status = domain.WorkspaceStatusError
        ws.Error = fmt.Sprintf("Failed to create container: %v", err)
        s.repo.Update(ws)
        return err
    }

    ws.ContainerID = containerID

    // 4. 启动容器
    if err := s.docker.StartContainer(ctx, containerID); err != nil {
        ws.Status = domain.WorkspaceStatusError
        ws.Error = fmt.Sprintf("Failed to start container: %v", err)
        s.repo.Update(ws)
        return err
    }

    // 5. 执行脚本（异步）
    go s.executeScripts(ctx, ws)

    return nil
}

// RestoreWorkspaces restores all workspaces on startup
func (s *WorkspaceService) RestoreWorkspaces(ctx context.Context) error {
    // 1. 清理所有旧的工作空间容器（防止异常退出残留）
    if err := s.CleanupContainers(ctx); err != nil {
        log.Printf("Warning: Failed to cleanup old containers: %v", err)
    }

    // 2. 加载所有工作空间配置
    workspaces, err := s.repo.List()
    if err != nil {
        return err
    }

    // 3. 重新创建所有工作空间
    for _, ws := range workspaces {
        log.Printf("Restoring workspace: %s", ws.Name)

        // 清空运行时字段
        ws.ContainerID = ""
        ws.Status = domain.WorkspaceStatusCreating
        ws.Error = ""

        // 重新创建容器并执行脚本
        if err := s.createAndStartWorkspace(ctx, ws); err != nil {
            log.Printf("Failed to restore workspace %s: %v", ws.Name, err)
            ws.Status = domain.WorkspaceStatusError
            ws.Error = err.Error()
        }

        s.repo.Update(ws)
    }

    return nil
}

// CleanupContainers removes all ViBox workspace containers
func (s *WorkspaceService) CleanupContainers(ctx context.Context) error {
    // 查找所有带有 vibox.workspace 标签的容器
    containers, err := s.docker.ListContainers(ctx, map[string]string{
        "label": "vibox.workspace",
    })
    if err != nil {
        return err
    }

    for _, container := range containers {
        log.Printf("Cleaning up old container: %s", container.ID)
        s.docker.StopContainer(ctx, container.ID)
        s.docker.RemoveContainer(ctx, container.ID)
    }

    return nil
}

// Shutdown gracefully shuts down the service and cleanup containers
func (s *WorkspaceService) Shutdown(ctx context.Context) error {
    log.Printf("Shutting down workspace service...")

    // 删除所有工作空间容器
    if err := s.CleanupContainers(ctx); err != nil {
        log.Printf("Warning: Failed to cleanup containers during shutdown: %v", err)
    }

    log.Printf("Workspace service shutdown complete")
    return nil
}
```

---

### 3. Handler 层：新增端点

```go
// internal/api/handler/workspace.go

// UpdatePorts updates workspace port mappings
func (h *WorkspaceHandler) UpdatePorts(c *gin.Context) {
    id := c.Param("id")

    var req struct {
        Ports map[string]string `json:"ports" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request body",
            "code":  "INVALID_REQUEST",
        })
        return
    }

    if err := h.service.UpdatePorts(c.Request.Context(), id, req.Ports); err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            c.JSON(http.StatusNotFound, gin.H{
                "error": "Workspace not found",
                "code":  "NOT_FOUND",
            })
            return
        }

        c.JSON(http.StatusInternalServerError, gin.H{
            "error": fmt.Sprintf("Failed to update ports: %v", err),
            "code":  "INTERNAL_ERROR",
        })
        return
    }

    ws, _ := h.service.GetWorkspace(id)
    c.JSON(http.StatusOK, ws)
}

// ResetWorkspace resets a workspace to initial state
func (h *WorkspaceHandler) ResetWorkspace(c *gin.Context) {
    id := c.Param("id")

    if err := h.service.ResetWorkspace(c.Request.Context(), id); err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            c.JSON(http.StatusNotFound, gin.H{
                "error": "Workspace not found",
                "code":  "NOT_FOUND",
            })
            return
        }

        c.JSON(http.StatusInternalServerError, gin.H{
            "error": fmt.Sprintf("Failed to reset workspace: %v", err),
            "code":  "INTERNAL_ERROR",
        })
        return
    }

    ws, _ := h.service.GetWorkspace(id)
    c.JSON(http.StatusOK, gin.H{
        "message": "Workspace reset successfully",
        "workspace": ws,
    })
}
```

---

### 4. 路由配置

```go
// internal/api/router.go

func SetupRouter(
    cfg *config.Config,
    workspaceHandler *handler.WorkspaceHandler,
    // ...
) *gin.Engine {
    r := gin.Default()

    // ... 其他中间件 ...

    // API 路由
    api := r.Group("/api", middleware.AuthMiddleware(cfg.APIToken))
    {
        // 工作空间管理
        api.POST("/workspaces", workspaceHandler.Create)
        api.GET("/workspaces", workspaceHandler.List)
        api.GET("/workspaces/:id", workspaceHandler.Get)
        api.DELETE("/workspaces/:id", workspaceHandler.Delete)

        // 新增：端口管理
        api.PUT("/workspaces/:id/ports", workspaceHandler.UpdatePorts)

        // 新增：重置工作空间
        api.POST("/workspaces/:id/reset", workspaceHandler.ResetWorkspace)
    }

    // ... 其他路由 ...

    return r
}
```

---

### 5. 启动流程

```go
// cmd/server/main.go

func main() {
    // 1. 加载配置
    cfg := config.Load()

    // 2. 初始化 Docker 客户端
    dockerClient, err := docker.NewClient(cfg.DockerHost)
    if err != nil {
        log.Fatalf("Failed to create Docker client: %v", err)
    }

    // 3. 初始化 Repository（会自动加载持久化数据）
    dataDir := getEnv("DATA_DIR", "./data")
    workspaceRepo := repository.NewWorkspaceRepository(dataDir)

    // 4. 初始化 Service
    dockerService := service.NewDockerService(dockerClient, cfg)
    workspaceService := service.NewWorkspaceService(workspaceRepo, dockerService)

    // 5. 恢复工作空间（新增）
    ctx := context.Background()
    if err := workspaceService.RestoreWorkspaces(ctx); err != nil {
        log.Printf("Warning: Failed to restore workspaces: %v", err)
    }

    // 6. 初始化 Handler 和路由
    // ...

    // 7. 设置信号处理（新增）
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // 8. 在 goroutine 中启动服务器
    srv := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: r,
    }

    go func() {
        log.Printf("Starting ViBox server on :%s", cfg.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    // 9. 等待退出信号（新增）
    <-sigChan
    log.Println("Received shutdown signal")

    // 10. 优雅关闭（新增）
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 关闭 HTTP 服务器
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }

    // 清理工作空间容器
    if err := workspaceService.Shutdown(shutdownCtx); err != nil {
        log.Printf("Workspace service shutdown error: %v", err)
    }

    log.Println("ViBox server stopped")
}
```

---

## 迁移方案

### 数据迁移

**从 v1.0.0 迁移到 v1.1.0**：

由于 v1.0.0 没有持久化，不存在旧数据，无需迁移。

**后续版本迁移**：

如果需要从 v1.1.0 迁移到 v1.2.0（假设数据格式变更）：

```go
// internal/repository/migration.go

func MigrateV1ToV2(oldData *PersistentDataV1) *PersistentDataV2 {
    newData := &PersistentDataV2{
        Version:   "1.2.0",
        Workspaces: make(map[string]*WorkspaceV2),
    }

    for id, ws := range oldData.Workspaces {
        newData.Workspaces[id] = &WorkspaceV2{
            ID:     ws.ID,
            Name:   ws.Name,
            // ... 转换字段 ...
        }
    }

    return newData
}
```

---

## Docker 部署变更

### docker-compose.yml

```yaml
version: '3.8'

services:
  vibox:
    image: ghcr.io/1percentsync/vibox:latest
    ports:
      - "${HOST_PORT:-3000}:3000"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - vibox-data:/data  # 新增：持久化数据卷
    environment:
      - API_TOKEN=${API_TOKEN:?API_TOKEN is required}
      - PORT=${PORT:-3000}
      - DATA_DIR=/data    # 新增：数据目录
      - DEFAULT_IMAGE=${DEFAULT_IMAGE:-ubuntu:22.04}
    restart: unless-stopped

volumes:
  vibox-data:  # 新增：数据卷定义
    driver: local
```

### Dockerfile

```dockerfile
# 无需修改，但需要确保 /data 目录存在
FROM alpine:latest

# ...

# 创建数据目录
RUN mkdir -p /data && chown vibox:vibox /data

# ...
```

---

## 测试计划

### 1. 端口映射测试

```bash
# 创建工作空间（带端口映射）
curl -X POST http://localhost:3000/api/workspaces \
  -H "X-ViBox-Token: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-ports",
    "image": "ubuntu:22.04",
    "ports": {
      "8080": "VS Code Server",
      "3000": "Web App"
    }
  }'

# 验证端口映射
curl http://localhost:3000/api/workspaces/ws-xxx \
  -H "X-ViBox-Token: $TOKEN" | jq '.ports'

# 更新端口映射
curl -X PUT http://localhost:3000/api/workspaces/ws-xxx/ports \
  -H "X-ViBox-Token: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ports": {
      "8080": "VS Code Server",
      "5432": "PostgreSQL"
    }
  }'
```

### 2. 容器重置测试

```bash
# 创建工作空间
WS_ID=$(curl -X POST http://localhost:3000/api/workspaces \
  -H "X-ViBox-Token: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-reset",
    "image": "ubuntu:22.04",
    "scripts": [{
      "name": "create-file",
      "content": "#!/bin/bash\necho hello > /tmp/test.txt",
      "order": 1
    }]
  }' | jq -r '.id')

# 等待创建完成
sleep 5

# 验证文件存在
docker exec $(docker ps -q -f label=vibox.workspace=$WS_ID) cat /tmp/test.txt
# 输出：hello

# 删除文件（模拟容器状态变更）
docker exec $(docker ps -q -f label=vibox.workspace=$WS_ID) rm /tmp/test.txt

# 重置工作空间
curl -X POST http://localhost:3000/api/workspaces/$WS_ID/reset \
  -H "X-ViBox-Token: $TOKEN"

# 等待重置完成
sleep 5

# 验证文件已恢复
docker exec $(docker ps -q -f label=vibox.workspace=$WS_ID) cat /tmp/test.txt
# 输出：hello
```

### 3. 持久化测试

```bash
# 创建工作空间（auto_restore=true）
curl -X POST http://localhost:3000/api/workspaces \
  -H "X-ViBox-Token: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-persist",
    "image": "ubuntu:22.04",
    "auto_restore": true
  }'

# 验证数据文件存在
cat ./data/workspaces.json

# 重启 ViBox 服务
docker-compose restart vibox

# 等待启动
sleep 10

# 验证工作空间已恢复
curl http://localhost:3000/api/workspaces \
  -H "X-ViBox-Token: $TOKEN" | jq '.[] | select(.name=="test-persist")'
```

---

## 影响范围

### 需要修改的文件

#### 新增文件
- `docs/BACKEND_ENHANCEMENTS.md` - 本文档

#### 需要更新的文档
- `docs/API_SPECIFICATION.md` - 添加新 API 端点
- `docs/PHASE1_BACKEND.md` - 更新架构说明
- `README.md` - 更新功能列表
- `PROJECT_ROADMAP.md` - 更新完成状态

#### 需要修改的代码文件
- `internal/domain/workspace.go` - 添加 Ports、AutoRestore 字段
- `internal/repository/workspace.go` - 实现持久化存储
- `internal/service/workspace.go` - 添加 UpdatePorts、ResetWorkspace、RestoreWorkspaces 方法
- `internal/service/docker.go` - 添加 ContainerExists、IsContainerRunning 方法
- `internal/api/handler/workspace.go` - 添加新端点 Handler
- `internal/api/router.go` - 注册新路由
- `cmd/server/main.go` - 添加启动时恢复逻辑
- `docker-compose.yml` - 添加数据卷
- `.env.example` - 添加 DATA_DIR 配置

---

## 版本计划

| 版本 | 功能 | 状态 |
|------|------|------|
| v1.0.0 | 基础后端功能 | ✅ 已完成 |
| v1.1.0 | 端口标签 + 容器重置 + 数据持久化 | 📝 规划中 |
| v1.2.0 | 前端界面 | 📍 进行中 |

---

## 总结

本增强方案为 ViBox 后端添加了三个重要功能：

1. **端口标签功能** - 前端可以显示快捷访问按钮，提升用户体验
2. **容器重置功能** - 允许用户快速恢复工作空间到初始状态
3. **数据持久化功能** - 主容器重启后自动恢复工作空间，提高可靠性

这些功能为第二阶段（前端开发）提供了更完善的后端支持，同时也为未来的扩展（如 VS Code Server 集成）打下了基础。

---

**下一步**：
1. 更新相关文档
2. 实现代码变更
3. 测试验证
4. 发布 v1.1.0

---

**作者**：Claude
**日期**：2025-11-10
**版本**：v1.0.0
