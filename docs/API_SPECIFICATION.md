# ViBox API 规范

> **版本**：v1.0.0
> **阶段**：第一阶段（Go 后端）

---

## 目录

1. [通用规范](#通用规范)
2. [鉴权机制](#鉴权机制)
3. [工作空间管理 API](#工作空间管理-api)
4. [WebSocket 终端 API](#websocket-终端-api)
5. [端口转发 API](#端口转发-api)
6. [错误处理](#错误处理)
7. [状态码说明](#状态码说明)

---

## 通用规范

### Base URL

```
http://localhost:3000
```

生产环境通过 Caddy 反向代理：
```
https://your-domain.com
```

### Content-Type

所有请求和响应使用 JSON 格式（WebSocket 和 Proxy 除外）：
```
Content-Type: application/json
```

### 时间格式

所有时间字段使用 ISO 8601 格式：
```
2025-11-10T12:00:00Z
```

### 通用响应头

```
Content-Type: application/json; charset=utf-8
X-Request-ID: {request-id}
```

---

## 鉴权机制

### Cookie 鉴权

ViBox 使用 **HTTP Cookie** 作为主要的身份验证方式。

#### 登录流程

1. 用户通过登录接口提交 API Token
2. 后端验证 token 并设置 Cookie
3. 后续所有请求自动携带 Cookie（浏览器行为）

```http
POST /api/auth/login
Content-Type: application/json

{
  "token": "your-api-token"
}
```

**响应**：
```http
HTTP/1.1 200 OK
Set-Cookie: vibox-token=your-api-token; Path=/; Max-Age=86400; HttpOnly; SameSite=Lax

{
  "message": "Login successful"
}
```

**Cookie 属性说明**：
- `Path=/`：全局有效，所有路径都可用
- `Max-Age=86400`：24小时有效期
- `HttpOnly`：防止 JavaScript 访问，增强安全性
- `SameSite=Lax`：CSRF 防护

#### 认证方式

**方式 1：Cookie（推荐）**

浏览器自动携带 Cookie，无需手动设置：
```http
GET /api/workspaces
Cookie: vibox-token=your-api-token
```

**方式 2：查询参数（仅 WebSocket）**

WebSocket 连接可以使用查询参数：
```http
ws://localhost:3000/ws/terminal/:id?token=your-api-token
```

**使用场景**：
- WebSocket 连接（浏览器 WebSocket API 会自动发送 Cookie，但某些工具可能不支持）
- 外部工具访问（如 curl、Postman）

**注意**：
- Cookie 是首选方式（浏览器自动处理）
- 查询参数仅用于 WebSocket 连接或外部工具
- Token 会出现在 URL 中，有日志泄露风险

#### 认证优先级

```
Cookie > Query Parameter (?token=)
```

后端会按以下顺序检查：
1. 检查 Cookie 中的 `vibox-token`
2. 检查查询参数 `?token=`

### 鉴权失败响应

**API 请求（JSON）**：
```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": "Unauthorized: invalid or missing authentication",
  "code": "UNAUTHORIZED"
}
```

**浏览器请求（HTML）**：
```http
HTTP/1.1 302 Found
Location: /login
```

浏览器请求（`Accept: text/html`）会被重定向到登录页。

### 登出流程

清除 Cookie 即可：

```http
POST /api/auth/logout
```

**响应**：
```http
HTTP/1.1 200 OK
Set-Cookie: vibox-token=; Path=/; Max-Age=-1

{
  "message": "Logout successful"
}
```

---

## 认证 API

### 1. 登录

验证 API Token 并设置认证 Cookie。

```http
POST /api/auth/login
Content-Type: application/json
```

#### 请求体

```json
{
  "token": "your-api-token"
}
```

**字段说明**：

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `token` | string | ✅ | API Token（配置在环境变量 API_TOKEN 中） |

#### 成功响应

```http
HTTP/1.1 200 OK
Set-Cookie: vibox-token=your-api-token; Path=/; Max-Age=86400; HttpOnly; SameSite=Lax
Content-Type: application/json

{
  "message": "Login successful"
}
```

#### 错误响应

**Token 无效**：
```http
HTTP/1.1 401 Unauthorized

{
  "error": "Invalid token",
  "code": "UNAUTHORIZED"
}
```

**请求格式错误**：
```http
HTTP/1.1 400 Bad Request

{
  "error": "Invalid request: token is required",
  "code": "INVALID_REQUEST"
}
```

---

### 2. 登出

清除认证 Cookie。

```http
POST /api/auth/logout
```

#### 成功响应

```http
HTTP/1.1 200 OK
Set-Cookie: vibox-token=; Path=/; Max-Age=-1
Content-Type: application/json

{
  "message": "Logout successful"
}
```

**注意**：登出接口无需认证，任何人都可以调用（只是清除 Cookie）。

---

## 工作空间管理 API

### 1. 创建工作空间

创建一个新的工作空间（Docker 容器）。

```http
POST /api/workspaces
Authorization: Bearer {token}
Content-Type: application/json
```

#### 请求体

```json
{
  "name": "my-workspace",
  "image": "ubuntu:22.04",
  "scripts": [
    {
      "name": "install-tools",
      "content": "#!/bin/bash\napt-get update && apt-get install -y curl git vim",
      "order": 1
    },
    {
      "name": "setup-user",
      "content": "#!/bin/bash\nuseradd -m -s /bin/bash developer",
      "order": 2
    }
  ],
  "ports": {
    "8080": "VS Code Server",
    "3000": "Web App"
  }
}
```

**字段说明**：

| 字段 | 类型 | 必需 | 默认值 | 说明 |
|------|------|------|--------|------|
| `name` | string | ✅ | - | 工作空间名称（唯一） |
| `image` | string | ❌ | `ubuntu:22.04` | Docker 镜像 |
| `scripts` | array | ❌ | `[]` | 初始化脚本列表 |
| `scripts[].name` | string | ✅ | - | 脚本名称 |
| `scripts[].content` | string | ✅ | - | 脚本内容（Bash） |
| `scripts[].order` | integer | ✅ | - | 执行顺序（从小到大） |
| `ports` | object | ❌ | `{}` | 端口标签映射（key=端口号，value=服务名） |

#### 成功响应

```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "id": "ws-a1b2c3d4",
  "name": "my-workspace",
  "container_id": "docker-abc123",
  "status": "creating",
  "created_at": "2025-11-10T12:00:00Z",
  "config": {
    "image": "ubuntu:22.04",
    "scripts": [
      {
        "name": "install-tools",
        "content": "#!/bin/bash\napt-get update && apt-get install -y curl git vim",
        "order": 1
      },
      {
        "name": "setup-user",
        "content": "#!/bin/bash\nuseradd -m -s /bin/bash developer",
        "order": 2
      }
    ]
  },
  "ports": {
    "8080": "VS Code Server",
    "3000": "Web App"
  }
}
```

**状态说明**：
- `creating` - 正在创建容器和执行脚本
- 脚本执行完成后自动更新为 `running`
- 脚本执行失败时更新为 `error`

#### 错误响应

**请求验证失败**：
```http
HTTP/1.1 400 Bad Request

{
  "error": "Invalid request: name is required",
  "code": "INVALID_REQUEST"
}
```

**Docker 操作失败**：
```http
HTTP/1.1 500 Internal Server Error

{
  "error": "Failed to create container: unable to pull image",
  "code": "DOCKER_ERROR"
}
```

---

### 2. 列出工作空间

获取所有工作空间列表。

```http
GET /api/workspaces
Authorization: Bearer {token}
```

#### 成功响应

```http
HTTP/1.1 200 OK
Content-Type: application/json

[
  {
    "id": "ws-a1b2c3d4",
    "name": "my-workspace",
    "container_id": "docker-abc123",
    "status": "running",
    "created_at": "2025-11-10T12:00:00Z",
    "updated_at": "2025-11-10T12:01:30Z",
    "config": {
      "image": "ubuntu:22.04",
      "scripts": [...]
    }
  },
  {
    "id": "ws-e5f6g7h8",
    "name": "test-workspace",
    "container_id": "docker-def456",
    "status": "stopped",
    "created_at": "2025-11-09T10:30:00Z",
    "updated_at": "2025-11-09T11:00:00Z",
    "config": {
      "image": "alpine:latest",
      "scripts": []
    }
  }
]
```

**空列表**：
```json
[]
```

---

### 3. 获取工作空间详情

获取单个工作空间的详细信息。

```http
GET /api/workspaces/:id
Authorization: Bearer {token}
```

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | string | 工作空间 ID |

#### 成功响应

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": "ws-a1b2c3d4",
  "name": "my-workspace",
  "container_id": "docker-abc123",
  "status": "running",
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-10T12:01:30Z",
  "config": {
    "image": "ubuntu:22.04",
    "scripts": [...]
  }
}
```

#### 错误响应

**工作空间不存在**：
```http
HTTP/1.1 404 Not Found

{
  "error": "Workspace not found",
  "code": "NOT_FOUND"
}
```

---

### 4. 删除工作空间

删除工作空间及其 Docker 容器。

```http
DELETE /api/workspaces/:id
Authorization: Bearer {token}
```

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | string | 工作空间 ID |

#### 成功响应

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "message": "Workspace deleted successfully",
  "id": "ws-a1b2c3d4"
}
```

#### 错误响应

**工作空间不存在**：
```http
HTTP/1.1 404 Not Found

{
  "error": "Workspace not found",
  "code": "NOT_FOUND"
}
```

**Docker 操作失败**：
```http
HTTP/1.1 500 Internal Server Error

{
  "error": "Failed to delete container: container is locked",
  "code": "DOCKER_ERROR"
}
```

---

### 5. 更新端口映射

更新工作空间的端口标签映射。

```http
PUT /api/workspaces/:id/ports
X-ViBox-Token: {token}
Content-Type: application/json
```

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | string | 工作空间 ID |

#### 请求体

```json
{
  "ports": {
    "8080": "VS Code Server",
    "3000": "Web App",
    "5432": "PostgreSQL"
  }
}
```

**字段说明**：

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `ports` | object | ✅ | 端口标签映射（key=端口号，value=服务名） |

#### 成功响应

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": "ws-a1b2c3d4",
  "name": "my-workspace",
  "ports": {
    "8080": "VS Code Server",
    "3000": "Web App",
    "5432": "PostgreSQL"
  },
  "updated_at": "2025-11-10T12:05:00Z"
}
```

#### 错误响应

**工作空间不存在**：
```http
HTTP/1.1 404 Not Found

{
  "error": "Workspace not found",
  "code": "NOT_FOUND"
}
```

**请求验证失败**：
```http
HTTP/1.1 400 Bad Request

{
  "error": "Invalid request: ports is required",
  "code": "INVALID_REQUEST"
}
```

---

### 6. 重置工作空间

重置工作空间到初始状态（删除旧容器，重新创建并执行脚本）。

```http
POST /api/workspaces/:id/reset
X-ViBox-Token: {token}
```

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | string | 工作空间 ID |

#### 功能说明

1. 停止并删除旧容器
2. 使用原始配置创建新容器
3. 重新执行所有初始化脚本
4. 保留工作空间 ID、配置和端口映射

**使用场景**：
- 脚本执行失败，需要重新运行
- 容器状态混乱，需要恢复干净环境
- 测试脚本，需要多次重置

#### 成功响应

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "message": "Workspace reset successfully",
  "workspace": {
    "id": "ws-a1b2c3d4",
    "name": "my-workspace",
    "container_id": "docker-new123",
    "status": "creating",
    "updated_at": "2025-11-10T12:10:00Z"
  }
}
```

**状态说明**：
- 重置后状态为 `creating`
- 脚本执行完成后自动更新为 `running`
- 脚本执行失败时更新为 `error`

#### 错误响应

**工作空间不存在**：
```http
HTTP/1.1 404 Not Found

{
  "error": "Workspace not found",
  "code": "NOT_FOUND"
}
```

**Docker 操作失败**：
```http
HTTP/1.1 500 Internal Server Error

{
  "error": "Failed to create container: unable to pull image",
  "code": "DOCKER_ERROR"
}
```

---

## WebSocket 终端 API

### 连接到终端

建立 WebSocket 连接到工作空间的终端。

```
ws://localhost:3000/ws/terminal/:id?token={token}
```

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | string | 工作空间 ID |

#### 查询参数

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `token` | string | ✅ | API Token |

### 消息协议

#### 客户端 → 服务器

**1. 用户输入**

```json
{
  "type": "input",
  "data": "ls -la\n"
}
```

**2. 终端大小调整**

```json
{
  "type": "resize",
  "cols": 80,
  "rows": 24
}
```

#### 服务器 → 客户端

**1. 终端输出**

```json
{
  "type": "output",
  "data": "total 48\ndrwxr-xr-x 2 root root 4096 Nov 10 12:00 .\ndrwxr-xr-x 3 root root 4096 Nov 10 11:59 ..\n"
}
```

**2. 错误消息**

```json
{
  "type": "error",
  "data": "Connection lost"
}
```

**3. 连接关闭**

```json
{
  "type": "close",
  "data": "Session terminated"
}
```

### 连接流程

```
客户端                        服务器
  │                            │
  │─────── WebSocket 升级 ────→│
  │←────── 101 Switching ──────│
  │                            │
  │──── {"type":"input",...} ─→│
  │←─── {"type":"output",...}──│
  │                            │
  │──── {"type":"resize",...}─→│
  │                            │
  │←─── {"type":"close",...}───│
  │────────── Close ──────────→│
```

### 错误响应

**工作空间不存在**：
```http
HTTP/1.1 404 Not Found

{
  "error": "Workspace not found",
  "code": "NOT_FOUND"
}
```

**容器未运行**：
```http
HTTP/1.1 400 Bad Request

{
  "error": "Container is not running",
  "code": "CONTAINER_NOT_RUNNING"
}
```

**鉴权失败**：
```http
HTTP/1.1 401 Unauthorized

{
  "error": "Unauthorized: invalid or missing token",
  "code": "UNAUTHORIZED"
}
```

### 客户端示例

#### JavaScript (xterm.js)

```javascript
// 创建终端
const term = new Terminal();
term.open(document.getElementById('terminal'));

// 连接 WebSocket
const ws = new WebSocket(`ws://localhost:3000/ws/terminal/${workspaceId}?token=${apiToken}`);

// 发送用户输入
term.onData(data => {
  ws.send(JSON.stringify({ type: 'input', data }));
});

// 接收终端输出
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'output') {
    term.write(msg.data);
  } else if (msg.type === 'error' || msg.type === 'close') {
    console.error(msg.data);
    ws.close();
  }
};

// 监听终端大小变化
term.onResize(({ cols, rows }) => {
  ws.send(JSON.stringify({ type: 'resize', cols, rows }));
});
```

#### Go (gorilla/websocket)

```go
// 连接 WebSocket
url := fmt.Sprintf("ws://localhost:3000/ws/terminal/%s?token=%s", workspaceID, apiToken)
ws, _, err := websocket.DefaultDialer.Dial(url, nil)
if err != nil {
    log.Fatal(err)
}
defer ws.Close()

// 发送输入
msg := TerminalMessage{Type: "input", Data: "ls -la\n"}
ws.WriteJSON(msg)

// 接收输出
var response TerminalMessage
for {
    if err := ws.ReadJSON(&response); err != nil {
        break
    }
    if response.Type == "output" {
        fmt.Print(response.Data)
    }
}
```

---

## 端口转发 API

### 访问容器内 HTTP 服务

将请求代理转发到容器内指定端口的 HTTP 服务。

**设计说明**：
- 端口访问采用**动态模式**：无需预先声明端口，可以访问容器的任意端口
- 工作空间容器**不会在宿主机上暴露端口**，所有访问都通过后端代理转发
- 如果端口没有服务监听，将返回 502 或 504 错误

**鉴权说明**：
- ⚠️ **必须通过认证** 才能访问端口转发功能
- **浏览器访问**：自动使用 Cookie 鉴权（登录后自动生效）
- **外部工具访问**：使用查询参数 `?token=` 或手动设置 Cookie
- ViBox 会自动移除 `vibox-token` Cookie，不会传递给容器
- 容器内应用的所有 header 和 Cookie 会被完整保留

```http
{METHOD} /forward/:id/:port/*path
Cookie: vibox-token={your-api-token}
```

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | string | 工作空间 ID |
| `port` | integer | 容器内端口号 |
| `path` | string | 请求路径（可选） |

#### 示例

**浏览器访问（推荐）**：

```javascript
// 登录后，浏览器自动携带 Cookie
window.open('/forward/ws-a1b2c3d4/8080/')
// Cookie 会自动发送，无需手动设置
```

**外部工具访问**：

```bash
# 方式1：使用查询参数（简单）
curl "http://localhost:3000/forward/ws-a1b2c3d4/8080/?token=your-token"

# 方式2：手动设置 Cookie
curl -H "Cookie: vibox-token=your-token" \
  http://localhost:3000/forward/ws-a1b2c3d4/8080/
```

**访问特定路径**：

```bash
curl "http://localhost:3000/forward/ws-a1b2c3d4/8080/api/users?token=your-token"
# 实际访问：http://{container-ip}:8080/api/users
```

**POST 请求**：

```bash
curl -X POST "http://localhost:3000/forward/ws-a1b2c3d4/3000/api/data?token=your-token" \
  -H "Content-Type: application/json" \
  -d '{"key": "value"}'
```

**同时使用容器应用的认证**：

```bash
# ViBox 层认证（Cookie） + 容器应用层认证（Header/Cookie）
curl "http://localhost:3000/forward/ws-a1b2c3d4/3000/api/protected?token=vibox-token" \
  -H "Authorization: Bearer app-user-token" \
  -H "Cookie: app-session=abc123"
# vibox-token Cookie 会被移除
# Authorization 和 app-session Cookie 会被转发给容器内应用
```

### 代理行为

#### 请求处理
- **请求头**：原样转发（除特殊处理的 header/cookie）
  - `Host`: 自动修改为容器 IP:端口
  - `Cookie: vibox-token`: **自动移除**，不会传递给容器
  - 其他 header/cookie: 完整保留，传递给容器应用
- **请求体**：原样转发
- **查询参数**：原样转发（包括 `?token=` 如果存在）

#### 响应处理
- **响应头**：原样返回
- **响应体**：原样返回
- **状态码**：原样返回

#### 自动添加的 Header
- `X-Forwarded-For`: 客户端 IP
- `X-Forwarded-Proto`: http 或 https

### 错误响应

**鉴权失败**：
```http
HTTP/1.1 401 Unauthorized

{
  "error": "Unauthorized: invalid or missing token",
  "code": "UNAUTHORIZED"
}
```

**工作空间不存在**：
```http
HTTP/1.1 404 Not Found

{
  "error": "Workspace not found",
  "code": "NOT_FOUND"
}
```

**容器未运行**：
```http
HTTP/1.1 400 Bad Request

{
  "error": "Container is not running",
  "code": "CONTAINER_NOT_RUNNING"
}
```

**端口未监听**：
```http
HTTP/1.1 502 Bad Gateway

{
  "error": "Failed to connect to container port",
  "code": "PROXY_ERROR"
}
```

---

## 错误处理

### 错误响应格式

所有错误响应遵循统一格式：

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional context (optional)"
  }
}
```

### 错误码列表

| 错误码 | HTTP 状态码 | 说明 |
|--------|------------|------|
| `UNAUTHORIZED` | 401 | Token 无效或缺失 |
| `FORBIDDEN` | 403 | 权限不足 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `INVALID_REQUEST` | 400 | 请求参数验证失败 |
| `DOCKER_ERROR` | 500 | Docker 操作失败 |
| `CONTAINER_NOT_RUNNING` | 400 | 容器未运行 |
| `PROXY_ERROR` | 502 | 端口转发失败 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |

### 错误示例

#### 请求验证失败

```http
HTTP/1.1 400 Bad Request

{
  "error": "Invalid request: name is required",
  "code": "INVALID_REQUEST",
  "details": {
    "field": "name",
    "constraint": "required"
  }
}
```

#### Docker 操作失败

```http
HTTP/1.1 500 Internal Server Error

{
  "error": "Failed to create container: unable to pull image ubuntu:99.99",
  "code": "DOCKER_ERROR",
  "details": {
    "image": "ubuntu:99.99",
    "reason": "pull access denied"
  }
}
```

#### 容器未运行

```http
HTTP/1.1 400 Bad Request

{
  "error": "Container is not running",
  "code": "CONTAINER_NOT_RUNNING",
  "details": {
    "workspace_id": "ws-a1b2c3d4",
    "status": "stopped"
  }
}
```

---

## 状态码说明

### HTTP 状态码

| 状态码 | 含义 | 使用场景 |
|--------|------|----------|
| `200 OK` | 成功 | GET, DELETE 成功 |
| `201 Created` | 已创建 | POST 创建成功 |
| `400 Bad Request` | 请求错误 | 参数验证失败、业务逻辑错误 |
| `401 Unauthorized` | 未授权 | Token 无效或缺失 |
| `403 Forbidden` | 禁止访问 | 权限不足（未来功能） |
| `404 Not Found` | 未找到 | 资源不存在 |
| `500 Internal Server Error` | 服务器错误 | Docker 错误、未预期错误 |
| `502 Bad Gateway` | 网关错误 | 端口转发失败 |

### 工作空间状态

| 状态 | 说明 | 终端可用 | 可执行操作 |
|------|------|---------|-----------|
| `creating` | 正在创建容器和执行脚本 | ❌ | 查询 |
| `running` | 容器运行中，一切正常 | ✅ | 查询、终端、端口、重置、删除 |
| `error` | 脚本执行失败，但容器仍在运行 | ✅ | 查询、终端（调试）、重置、删除 |
| `failed` | 容器创建/启动失败或已停止 | ❌ | 查询、重置、删除 |

**状态说明**：
- **`creating`**：容器正在创建中或初始化脚本正在执行
- **`running`**：工作空间正常运行，所有功能可用
- **`error`**：初始化脚本执行失败，但容器仍在运行，可以通过终端调试
- **`failed`**：容器创建失败、启动失败或运行后停止（总是异常情况）

**重要**：
- `error` 状态下终端仍可用，方便用户调试脚本问题
- `failed` 状态下容器不可访问，只能重置或删除

---

## 完整示例流程

### 场景：登录并创建工作空间访问终端

#### 0. 登录（设置 Cookie）

```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"token": "my-secret-token"}'
```

**响应**：
```json
{
  "message": "Login successful"
}
```

Cookie 已保存到 `cookies.txt`，后续请求自动使用。

#### 1. 创建工作空间

```bash
curl -X POST http://localhost:3000/api/workspaces \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "name": "dev-env",
    "image": "ubuntu:22.04",
    "scripts": [
      {
        "name": "install-node",
        "content": "#!/bin/bash\ncurl -fsSL https://deb.nodesource.com/setup_20.x | bash -\napt-get install -y nodejs",
        "order": 1
      }
    ]
  }'
```

**响应**：
```json
{
  "id": "ws-xyz789",
  "name": "dev-env",
  "status": "creating",
  ...
}
```

#### 2. 轮询工作空间状态

```bash
curl http://localhost:3000/api/workspaces/ws-xyz789 \
  -b cookies.txt
```

**等待 status 变为 `running`**

#### 3. 连接到终端

```javascript
const ws = new WebSocket('ws://localhost:3000/ws/terminal/ws-xyz789?token=my-secret-token');

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  console.log(msg.data);
};

ws.send(JSON.stringify({
  type: 'input',
  data: 'node --version\n'
}));

// 输出：v20.x.x
```

#### 4. 在容器内启动 HTTP 服务

```bash
# 通过终端执行
echo "const http = require('http'); http.createServer((req, res) => res.end('Hello')).listen(3000);" > server.js
node server.js &
```

#### 5. 通过端口转发访问

```bash
curl "http://localhost:3000/forward/ws-xyz789/3000/?token=my-secret-token"
# 输出：Hello
```

#### 6. 删除工作空间

```bash
curl -X DELETE http://localhost:3000/api/workspaces/ws-xyz789 \
  -b cookies.txt
```

**响应**：
```json
{
  "message": "Workspace deleted successfully",
  "id": "ws-xyz789"
}
```

---

## 版本历史

- **v1.0.0** (2025-11-10): 初始版本，第一阶段 API 规范

---

## 参考

- [第一阶段开发计划](./PHASE1_BACKEND.md)
- [任务拆分文档](./PHASE1_TASK_BREAKDOWN.md)
- [项目路线图](../PROJECT_ROADMAP.md)
