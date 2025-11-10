# WebSSH 实现原理详解

## 概述

WebSSH 是一种在浏览器中通过 WebSocket 实现的终端访问技术，允许用户无需本地 SSH 客户端即可访问远程服务器或容器的 Shell。

## 核心概念

### 1. 伪终端（PTY - Pseudo Terminal）

**什么是 PTY？**

PTY 是一对虚拟字符设备，包括：
- **Master 端**：程序（如我们的 Go 后端）读写的一端
- **Slave 端**：Shell 进程读写的一端，认为自己在真实终端中运行

**为什么需要 PTY？**

如果直接连接 stdin/stdout，Shell 会检测到不是终端（`isatty()` 返回 false），导致：
- 没有交互式提示符
- 不支持终端控制序列（颜色、光标移动等）
- 无法调整窗口大小
- 很多交互式程序无法正常工作

### 2. WebSocket

**为什么用 WebSocket 而不是 HTTP？**

- HTTP 是单向请求-响应模型，不适合双向实时通信
- WebSocket 提供全双工通信通道
- 低延迟，适合终端交互
- 浏览器原生支持

### 3. xterm.js

**前端终端模拟器**

- 在浏览器中模拟完整的终端（VT100/xterm 兼容）
- 处理 ANSI 转义序列（颜色、格式、光标控制）
- 捕获用户键盘输入
- 通过 WebSocket 与后端通信

---

## ViBox 中 WebSSH 的实现原理

### 架构流程图

```
┌────────────────────────────────────────────────────────────┐
│                    用户浏览器                               │
│  ┌──────────────────────────────────────────────────┐     │
│  │              xterm.js                            │     │
│  │  - 渲染终端界面                                  │     │
│  │  - 捕获键盘输入                                  │     │
│  │  - 显示输出内容                                  │     │
│  └──────────┬────────────────────────┬──────────────┘     │
│             │ 用户输入               │ 服务端输出         │
└─────────────┼────────────────────────┼────────────────────┘
              │ WebSocket              │
              ▼                        ▲
┌─────────────────────────────────────────────────────────────┐
│              Go 后端 WebSocket Handler                      │
│                                                              │
│  ┌────────────────────────────────────────────────┐        │
│  │  1. 升级 HTTP 连接为 WebSocket                 │        │
│  │  2. 根据 workspaceId 获取容器信息              │        │
│  │  3. 使用 Docker SDK ExecCreate 创建 exec 实例   │        │
│  │  4. 创建 PTY (creack/pty)                      │        │
│  │  5. 双向数据转发                               │        │
│  └────────────────────────────────────────────────┘        │
│                                                              │
│  数据流：                                                    │
│  WebSocket ─→ PTY Master ─→ Exec 进程 ─→ 容器 Shell        │
│  WebSocket ←─ PTY Master ←─ Exec 进程 ←─ 容器 Shell        │
└─────────────────────────┼───────────────────────────────────┘
                          │ Docker API
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Docker 容器                               │
│  ┌────────────────────────────────────────────────┐        │
│  │  /bin/bash (或其他 Shell)                      │        │
│  │  - 运行在 PTY Slave 端                         │        │
│  │  - 接收命令，执行，返回输出                    │        │
│  └────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

---

## 详细实现步骤

### 步骤 1: 前端发起 WebSocket 连接

```javascript
// 前端伪代码
const ws = new WebSocket('ws://domain.com/ws/terminal/workspace-123');
const term = new Terminal();

// 用户输入发送到后端
term.onData(data => {
  ws.send(JSON.stringify({ type: 'input', data: data }));
});

// 接收后端输出并显示
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'output') {
    term.write(msg.data);
  }
};

// 窗口大小变化时通知后端
term.onResize(({ cols, rows }) => {
  ws.send(JSON.stringify({ type: 'resize', cols, rows }));
});
```

### 步骤 2: Go 后端处理 WebSocket 连接

```go
// 后端伪代码
func HandleTerminal(c *gin.Context) {
    workspaceId := c.Param("workspaceId")

    // 1. 升级为 WebSocket
    ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer ws.Close()

    // 2. 获取容器信息
    container := GetWorkspace(workspaceId)

    // 3. 在容器中创建 exec 实例（执行 /bin/bash）
    exec, err := dockerClient.ContainerExecCreate(ctx, container.ID, types.ExecConfig{
        Cmd:          []string{"/bin/bash"},
        AttachStdin:  true,
        AttachStdout: true,
        AttachStderr: true,
        Tty:          true, // 重要！启用 TTY
    })

    // 4. 连接到 exec 实例
    hijackedConn, err := dockerClient.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{
        Tty: true,
    })
    defer hijackedConn.Close()

    // 5. 启动双向数据传输
    go func() {
        // WebSocket → 容器
        for {
            var msg Message
            ws.ReadJSON(&msg)

            switch msg.Type {
            case "input":
                hijackedConn.Conn.Write([]byte(msg.Data))
            case "resize":
                // 调整终端大小
                dockerClient.ContainerExecResize(ctx, exec.ID, types.ResizeOptions{
                    Height: msg.Rows,
                    Width:  msg.Cols,
                })
            }
        }
    }()

    // 容器 → WebSocket
    buf := make([]byte, 1024)
    for {
        n, err := hijackedConn.Reader.Read(buf)
        if err != nil {
            break
        }
        ws.WriteJSON(Message{
            Type: "output",
            Data: string(buf[:n]),
        })
    }
}
```

### 步骤 3: Docker 容器执行 Shell

容器内部发生的事情：
1. Docker daemon 在容器的命名空间内启动 `/bin/bash`
2. Bash 检测到 TTY 环境（因为我们设置了 `Tty: true`）
3. Bash 显示交互式提示符（如 `root@container-id:/# `）
4. 用户输入的命令通过 PTY 传递到 Bash
5. Bash 执行命令，输出通过 PTY 返回

---

## 关键技术点

### 1. 如何连接到容器？

我们使用 **Docker Exec API** 而不是 SSH：

**方法对比**：

| 方法 | 优势 | 劣势 |
|------|------|------|
| **Docker Exec** | 无需在容器内运行 SSH 服务，更轻量 | 需要访问 Docker Socket |
| SSH | 标准协议 | 需要在容器内安装配置 SSH，增加镜像体积 |

**ViBox 选择 Docker Exec**：
- 容器无需安装 SSH
- 更简单，更安全（不暴露 SSH 端口）
- 直接通过 Docker API 管理

### 2. TTY 标志的重要性

```go
// 创建 Exec 时必须设置
ExecConfig{
    Tty: true,  // 关键！
    // ...
}
```

如果 `Tty: false`：
- Shell 认为在非交互式环境
- 不显示提示符
- 很多交互式程序无法工作（如 vim, top）

### 3. 终端大小调整

当用户调整浏览器窗口大小时：

```
浏览器 → xterm.js 检测到窗口变化
       ↓
     WebSocket 发送 resize 消息 { cols: 80, rows: 24 }
       ↓
     Go 后端调用 ContainerExecResize()
       ↓
     容器内 Shell 收到 SIGWINCH 信号
       ↓
     Shell 重新调整输出格式
```

### 4. 数据编码

终端数据使用 **UTF-8** 编码，包含：
- 可打印字符（如 `a`, `中`）
- 控制字符（如 `\n`, `\r`, `\x1b`）
- ANSI 转义序列（如 `\x1b[31m` 表示红色）

xterm.js 能够正确解析和渲染这些序列。

---

## 完整数据流示例

### 用户输入 `ls -la` 并按回车

```
1. 用户在浏览器输入
   ┌──────────────────────┐
   │ xterm.js             │
   │ 捕获键盘事件：       │
   │ 'l', 's', ' ', '-', 'l', 'a', '\r' │
   └──────────┬───────────┘
              │
2. 通过 WebSocket 发送
   {"type": "input", "data": "ls -la\r"}
              │
              ▼
   ┌──────────────────────┐
   │ Go WebSocket Handler │
   │ 接收消息             │
   └──────────┬───────────┘
              │
3. 写入 Docker Exec 连接
   hijackedConn.Write("ls -la\r")
              │
              ▼
   ┌──────────────────────┐
   │ Docker 容器          │
   │ /bin/bash 接收输入   │
   │ 执行 ls -la          │
   │ 生成输出：           │
   │ total 48             │
   │ drwxr-xr-x  5 root... │
   │ ...                  │
   └──────────┬───────────┘
              │
4. 输出返回
   hijackedConn.Read() → "total 48\ndrwxr-xr-x..."
              │
              ▼
   ┌──────────────────────┐
   │ Go WebSocket Handler │
   │ 转发输出             │
   └──────────┬───────────┘
              │
5. WebSocket 发送到前端
   {"type": "output", "data": "total 48\n..."}
              │
              ▼
   ┌──────────────────────┐
   │ xterm.js             │
   │ 渲染输出             │
   └──────────────────────┘
```

---

## 会话管理

### 连接生命周期

```
1. 用户打开终端页面
   → WebSocket 连接建立
   → 创建 Docker Exec 实例
   → 启动 /bin/bash
   → 双向数据传输

2. 用户关闭页面或网络断开
   → WebSocket 连接断开
   → Go 检测到连接关闭
   → 关闭 Docker Exec 连接
   → 容器内 Bash 进程退出

3. 多个终端会话
   → 每个 WebSocket 连接对应一个独立的 Exec 实例
   → 可以同时打开多个终端访问同一个容器
   → 各个会话互不影响
```

### 会话清理

```go
// 伪代码
defer func() {
    // 确保资源清理
    hijackedConn.Close()
    ws.Close()
    // Docker Exec 会在连接关闭后自动清理
}()
```

---

## 安全考虑

### 1. 权限控制

- 后端需要验证用户是否有权访问该工作空间
- WebSocket 连接时检查认证信息

### 2. 资源限制

- 限制每个用户的并发终端会话数
- 设置会话超时（如 1 小时无活动自动断开）

### 3. 命令审计

可选：记录用户执行的命令用于审计
```go
// 记录输入
if msg.Type == "input" {
    AuditLog(workspaceId, msg.Data)
}
```

---

## 与传统 SSH 的对比

| 特性 | WebSSH (ViBox) | 传统 SSH |
|------|---------------|---------|
| 访问方式 | 浏览器 | SSH 客户端 |
| 容器要求 | 无需 SSH 服务 | 需要安装配置 SSH |
| 端口暴露 | 不需要 | 需要暴露 22 端口 |
| 认证方式 | Web 认证 + Docker API | SSH 密钥/密码 |
| 防火墙友好 | ✅ HTTPS/WSS | 可能被阻止 |
| 复杂度 | 较低 | 较高 |

---

## 实现时的注意事项

### 1. WebSocket 升级

确保正确设置 HTTP 头：
```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // 生产环境需要严格检查
    },
}
```

### 2. 错误处理

```go
// 容器可能不存在
// Exec 可能创建失败
// 网络可能中断
// 需要优雅处理所有错误
```

### 3. 字符编码

确保使用 UTF-8，支持中文等多字节字符。

### 4. 性能优化

- 使用缓冲读写
- 避免频繁的小数据传输
- 考虑压缩（WebSocket 支持）

---

## 总结

WebSSH 在 ViBox 中的实现流程：

1. **前端**：xterm.js 在浏览器中模拟终端，通过 WebSocket 通信
2. **后端**：Go 服务处理 WebSocket，使用 Docker Exec API 连接容器
3. **容器**：运行 Shell 进程，通过 TTY 提供交互式环境
4. **双向传输**：用户输入 → WebSocket → Docker Exec → Shell，输出反向流动

**核心优势**：
- 无需在容器内安装 SSH
- 用户只需浏览器即可访问
- 通过 Docker API 统一管理
- 部署简单，安全可控

相关代码实现见：[Go 后端架构设计](./BACKEND_ARCHITECTURE.md)
