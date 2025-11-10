# ViBox 后端开发计划

## 开发目标

实现 ViBox MVP 版本的 Go 后端，提供以下核心功能：
1. ✅ Docker 容器管理（创建、启动、停止、删除）
2. ✅ 自定义脚本执行
3. ✅ WebSSH 终端访问
4. ✅ HTTP 端口转发

---

## 开发阶段

### Phase 0: 环境准备（1 天）

#### 目标
搭建开发环境和项目骨架

#### 任务清单
- [ ] 安装 Go 1.21+
- [ ] 安装 Docker
- [ ] 初始化 Go 模块：`go mod init vibox`
- [ ] 创建项目目录结构（参考架构文档）
- [ ] 配置 `.gitignore`
- [ ] 安装核心依赖：
  ```bash
  go get github.com/gin-gonic/gin
  go get github.com/docker/docker
  go get github.com/gorilla/websocket
  go get github.com/google/uuid
  ```

#### 验收标准
- 项目结构清晰
- 依赖安装成功
- 可以运行 `go build ./cmd/server` 成功编译

---

### Phase 1: 基础架构（2-3 天）

#### 目标
搭建项目基础框架和 Docker 集成

#### 任务清单

**1.1 配置管理**
- [ ] 实现 `internal/config/config.go`
- [ ] 支持环境变量配置
- [ ] 定义默认配置

**1.2 工具函数**
- [ ] 实现 `pkg/utils/id.go`（UUID 生成）
- [ ] 实现 `pkg/utils/logger.go`（日志工具）

**1.3 领域模型**
- [ ] 实现 `internal/domain/workspace.go`
  - Workspace 结构体
  - WorkspaceConfig 结构体
  - Script 结构体
  - ExposedPort 结构体
  - 请求/响应结构体

**1.4 Docker 服务**
- [ ] 实现 `internal/service/docker.go`
  - NewDockerService（连接 Docker）
  - CreateContainer
  - StartContainer
  - StopContainer
  - RemoveContainer
  - GetContainerIP

**1.5 存储层**
- [ ] 实现 `internal/repository/workspace.go`
  - WorkspaceRepository 接口
  - InMemoryWorkspaceRepository 实现

**1.6 基础 API**
- [ ] 实现 `internal/api/middleware/logger.go`
- [ ] 实现 `internal/api/middleware/recovery.go`
- [ ] 实现 `internal/api/middleware/cors.go`
- [ ] 实现 `internal/api/router.go`（基础路由）

**1.7 程序入口**
- [ ] 实现 `cmd/server/main.go`
  - 依赖注入
  - 路由配置
  - 服务启动

#### 测试任务
- [ ] 测试 Docker 连接
- [ ] 测试容器创建/启动/停止/删除
- [ ] 测试基础 API 响应

#### 验收标准
- 服务可以启动在 `:3000`
- 可以通过 Docker SDK 操作容器
- 基础中间件工作正常
- 日志输出清晰

---

### Phase 2: 工作空间管理（3-4 天）

#### 目标
实现完整的工作空间 CRUD 功能

#### 任务清单

**2.1 工作空间服务**
- [ ] 实现 `internal/service/workspace.go`
  - NewWorkspaceService
  - CreateWorkspace（异步创建容器）
  - GetWorkspace
  - ListWorkspaces
  - DeleteWorkspace
  - executeScripts（执行初始化脚本）

**2.2 工作空间 API Handler**
- [ ] 实现 `internal/api/handler/workspace.go`
  - Create（POST /api/workspaces）
  - Get（GET /api/workspaces/:id）
  - List（GET /api/workspaces）
  - Delete（DELETE /api/workspaces/:id）

**2.3 脚本执行**
- [ ] 实现脚本复制到容器
- [ ] 实现按顺序执行脚本
- [ ] 实现脚本执行日志收集
- [ ] 实现脚本执行错误处理

**2.4 状态管理**
- [ ] 实现工作空间状态更新（creating → running/error）
- [ ] 实现容器健康检查

#### 测试任务
- [ ] 测试创建工作空间（不含脚本）
- [ ] 测试创建工作空间（包含脚本）
- [ ] 测试查询工作空间
- [ ] 测试删除工作空间
- [ ] 测试脚本执行顺序
- [ ] 测试脚本执行失败处理

#### API 测试示例
```bash
# 创建工作空间
curl -X POST http://localhost:3000/api/workspaces \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-workspace",
    "image": "ubuntu:22.04",
    "scripts": [
      {
        "name": "install-tools",
        "content": "#!/bin/bash\napt-get update && apt-get install -y curl",
        "order": 1
      }
    ]
  }'

# 列出工作空间
curl http://localhost:3000/api/workspaces

# 获取工作空间
curl http://localhost:3000/api/workspaces/{id}

# 删除工作空间
curl -X DELETE http://localhost:3000/api/workspaces/{id}
```

#### 验收标准
- 可以创建工作空间
- 容器自动启动
- 脚本按顺序执行成功
- 可以查询工作空间状态
- 可以删除工作空间及容器

---

### Phase 3: WebSSH 终端（4-5 天）

#### 目标
实现浏览器访问容器终端

#### 任务清单

**3.1 终端服务**
- [ ] 实现 `internal/service/terminal.go`
  - NewTerminalService
  - CreateSession（创建终端会话）
  - TerminalSession 结构体
  - handleInput（WebSocket → 容器）
  - handleOutput（容器 → WebSocket）
  - 终端大小调整（resize）

**3.2 WebSocket Handler**
- [ ] 实现 `internal/api/handler/terminal.go`
  - Connect（WebSocket 连接处理）
  - WebSocket 升级
  - 会话创建
  - 错误处理

**3.3 Docker Exec 集成**
- [ ] 实现在容器中创建 Exec 实例
- [ ] 实现 Exec Attach（hijacked 连接）
- [ ] 实现 TTY 支持
- [ ] 实现终端 resize

**3.4 消息协议**
- [ ] 定义 WebSocket 消息格式：
  ```json
  {
    "type": "input",
    "data": "ls -la\n"
  }
  ```
  ```json
  {
    "type": "output",
    "data": "total 48\ndrwxr-xr-x..."
  }
  ```
  ```json
  {
    "type": "resize",
    "cols": 80,
    "rows": 24
  }
  ```

**3.5 会话管理**
- [ ] 实现会话清理（连接断开时）
- [ ] 实现会话超时机制（可选）
- [ ] 实现并发会话管理

#### 测试任务
- [ ] 测试 WebSocket 连接建立
- [ ] 测试终端输入/输出
- [ ] 测试终端大小调整
- [ ] 测试多个并发会话
- [ ] 测试会话断开清理
- [ ] 使用 `websocat` 或浏览器测试

#### 测试工具
```bash
# 使用 websocat 测试 WebSocket
websocat ws://localhost:3000/ws/terminal/{workspace-id}

# 发送输入
{"type":"input","data":"ls\n"}

# 应该收到输出
{"type":"output","data":"..."}
```

#### 验收标准
- WebSocket 连接成功
- 可以在终端执行命令
- 可以看到实时输出
- 支持交互式程序（如 vim）
- 终端大小调整正常
- 连接断开后资源清理

---

### Phase 4: HTTP 端口转发（2-3 天）

#### 目标
实现容器内 HTTP 服务访问

#### 任务清单

**4.1 代理服务**
- [ ] 实现 `internal/service/proxy.go`
  - NewProxyService
  - ProxyRequest（反向代理实现）
  - 使用 `httputil.ReverseProxy`

**4.2 代理 Handler**
- [ ] 实现 `internal/api/handler/proxy.go`
  - Forward（处理转发请求）
  - 路径参数解析（workspace-id, port）
  - 错误处理

**4.3 动态路由**
- [ ] 实现 `/forward/:id/:port/*path` 路由
- [ ] 实现路径重写（去除前缀）
- [ ] 实现 WebSocket 代理（如果需要）

**4.4 容器 IP 获取**
- [ ] 实现获取容器内部 IP
- [ ] 缓存容器 IP（可选优化）

#### 测试任务
- [ ] 在容器内启动简单 HTTP 服务
  ```bash
  # 在容器内执行
  python3 -m http.server 8080
  ```
- [ ] 测试通过前端访问：
  ```bash
  curl http://localhost:3000/forward/{workspace-id}/8080/
  ```
- [ ] 测试不同路径
- [ ] 测试 POST 请求
- [ ] 测试静态文件服务

#### 验收标准
- 可以访问容器内 HTTP 服务
- 路径正确转发
- POST/PUT 等请求正常
- 静态文件可以加载
- 错误处理正确（容器不存在、端口未开放等）

---

### Phase 5: 完善与优化（2-3 天）

#### 目标
完善功能、优化性能、增强稳定性

#### 任务清单

**5.1 错误处理**
- [ ] 统一错误响应格式
- [ ] 完善各模块错误处理
- [ ] 添加详细错误日志

**5.2 日志优化**
- [ ] 结构化日志（使用 `logrus` 或 `zap`）
- [ ] 请求日志
- [ ] 容器操作日志
- [ ] 终端会话日志

**5.3 资源限制**
- [ ] 设置容器资源限制（CPU、内存）
- [ ] 限制并发会话数
- [ ] 实现会话超时

**5.4 安全加固**
- [ ] CORS 配置优化
- [ ] WebSocket Origin 检查
- [ ] 添加基础认证（可选）

**5.5 性能优化**
- [ ] 连接池优化
- [ ] 缓存优化
- [ ] 减少 Docker API 调用

**5.6 文档**
- [ ] API 文档（Swagger/OpenAPI）
- [ ] 部署文档
- [ ] 使用说明

#### 验收标准
- 错误信息清晰
- 日志完整可追踪
- 资源使用合理
- 无明显性能瓶颈

---

### Phase 6: 容器化部署（1-2 天）

#### 目标
Docker 化后端服务

#### 任务清单

**6.1 Dockerfile**
- [ ] 编写多阶段构建 Dockerfile
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

**6.2 Docker Compose**
- [ ] 编写 `docker-compose.yml`
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

**6.3 配置优化**
- [ ] 环境变量配置
- [ ] 容器网络配置
- [ ] 数据持久化（如果需要）

#### 测试任务
- [ ] 构建镜像
- [ ] 启动容器
- [ ] 测试所有功能
- [ ] 测试容器重启

#### 验收标准
- 镜像构建成功
- 容器启动正常
- 所有功能正常工作
- 配置通过环境变量控制

---

## 时间估算

| 阶段 | 预计时间 | 说明 |
|------|---------|------|
| Phase 0: 环境准备 | 1 天 | 熟悉 Go、Docker |
| Phase 1: 基础架构 | 2-3 天 | 核心框架搭建 |
| Phase 2: 工作空间管理 | 3-4 天 | 包含脚本执行 |
| Phase 3: WebSSH 终端 | 4-5 天 | 核心功能，需仔细测试 |
| Phase 4: HTTP 端口转发 | 2-3 天 | 相对简单 |
| Phase 5: 完善与优化 | 2-3 天 | 稳定性提升 |
| Phase 6: 容器化部署 | 1-2 天 | 部署准备 |
| **总计** | **15-21 天** | 约 3-4 周 |

**注**：时间估算基于每天 4-6 小时有效开发时间。

---

## 开发建议

### 1. 迭代开发
- 每完成一个 Phase，进行充分测试
- 不要等到最后才测试集成

### 2. 版本控制
- 每个 Phase 完成后创建 Git tag
- 保持提交历史清晰

### 3. 测试驱动
- 编写单元测试（至少覆盖核心逻辑）
- 手动测试 API 和 WebSocket

### 4. 文档同步
- 代码注释清晰
- 更新 API 文档

### 5. 问题记录
- 遇到问题记录到 GitHub Issues
- 技术决策记录到文档

---

## 测试策略

### 单元测试
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/service

# 带覆盖率
go test -cover ./...
```

### 集成测试
- 使用 Docker 启动测试容器
- 测试完整的工作流程

### 手动测试
- 使用 `curl` 测试 API
- 使用浏览器/websocat 测试 WebSocket
- 使用实际场景测试（创建工作空间 → 终端访问 → 端口转发）

---

## 风险与应对

### 风险 1: Docker API 不熟悉
**应对**：
- 先阅读官方文档和示例
- 参考 Portainer 等开源项目

### 风险 2: WebSocket 实现复杂
**应对**：
- 先实现简单的 echo 测试
- 参考 gorilla/websocket 官方示例

### 风险 3: 终端交互问题
**应对**：
- 确保设置 `Tty: true`
- 测试各种终端程序（vim、top 等）

### 风险 4: 时间超出预期
**应对**：
- 优先实现核心功能
- 优化功能可以后续迭代

---

## 下一步行动

1. **立即开始**：Phase 0 环境准备
2. **第一个里程碑**：完成 Phase 1，能创建和管理容器
3. **核心里程碑**：完成 Phase 3，实现 WebSSH
4. **MVP 完成**：完成 Phase 4，所有核心功能就绪

---

## 参考资源

### Go 学习资源
- [Go 官方文档](https://go.dev/doc/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)

### Docker SDK
- [Docker SDK for Go](https://docs.docker.com/engine/api/sdk/)
- [GitHub: docker/docker](https://github.com/moby/moby)

### WebSocket
- [gorilla/websocket](https://github.com/gorilla/websocket)
- [WebSocket 协议 RFC 6455](https://tools.ietf.org/html/rfc6455)

### 相关项目参考
- [Portainer](https://github.com/portainer/portainer)
- [Wetty](https://github.com/butlerx/wetty)
- [ttyd](https://github.com/tsl0922/ttyd)

准备好开始了吗？我可以帮你创建项目初始结构！
