# Terminal功能增强总结

## 已完成的修复

### 1. 方向键导航修复 ✅

**问题**: 方向键无法在terminal中导航命令历史

**根本原因**: 后端使用 `/bin/sh` 而不是 `/bin/bash`，sh不支持readline功能

**修复方案**:
- 修改 `internal/service/terminal.go:69-99`
- 优先检测并使用 `/bin/bash`
- 如果bash不存在，fallback到 `/bin/sh`

**效果**:
- ✅ 方向键上下可以浏览命令历史
- ✅ 左右方向键可以编辑命令
- ✅ Home/End等键也正常工作

---

### 2. 同浏览器页面切换状态保持 ✅

**问题**: 切换到其他页面再回来，terminal内容全部清空

**根本原因**:
- Terminal组件unmount时销毁terminal实例
- WebSocket连接被关闭
- 重新mount时创建全新实例

**修复方案**:
- 创建全局terminal store (`stores/terminals.ts`)
- Terminal实例保存在Jotai全局状态中
- WebSocket连接也保存在全局状态
- 组件unmount时不销毁，只是从DOM分离
- 组件re-mount时复用已有实例

**效果**:
- ✅ 切换页面后terminal内容保留
- ✅ 命令历史完整保存
- ✅ WebSocket连接保持
- ✅ 滚动位置保留

---

## 跨浏览器Terminal持久化方案

### 当前实现的限制

**现状**: Terminal实例保存在**前端内存**中（Jotai store）

**限制**:
- ✅ 同一浏览器tab切换页面：内容保留
- ❌ 刷新页面：内容丢失
- ❌ 新开tab：无法访问旧terminal
- ❌ 换浏览器：无法访问旧terminal
- ❌ 换设备：无法访问旧terminal

### 跨浏览器持久化需求分析

你提到的需求："换浏览器登录，点击terminal应该还是会保留"

这意味着需要：
1. **服务端保存terminal会话**
2. **跨客户端共享terminal状态**
3. **terminal历史回放**

---

## 解决方案：使用tmux实现服务端持久化

### 方案概述

使用 **tmux** (Terminal Multiplexer) 在容器内管理terminal会话：

```
用户A (浏览器1) ──┐
                   ├──> ViBox Backend ──> tmux session (在容器内)
用户A (浏览器2) ──┘
```

### 架构设计

```
┌─────────────────────────────────────────────────────────┐
│  Client (任何浏览器/设备)                                 │
│  - React + xterm.js                                     │
│  - WebSocket 连接                                        │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  ViBox Backend (Go)                                     │
│  - WebSocket Handler                                    │
│  - Session Manager                                      │
│    • 检查tmux会话是否存在                                │
│    • 创建新会话或attach到已有会话                        │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Docker Container                                       │
│  ┌───────────────────────────────────────────────────┐ │
│  │  tmux session: "workspace-{id}"                   │ │
│  │  - bash shell                                     │ │
│  │  - 命令历史                                        │ │
│  │  - 输出缓冲区                                      │ │
│  │  - 保持运行（即使无客户端连接）                     │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### 实现步骤

#### 1. 修改容器镜像（确保tmux已安装）

工作空间容器需要预装tmux：

```dockerfile
# 在工作空间创建时的初始化脚本中
apt-get update && apt-get install -y tmux
```

或在ubuntu:rolling基础上创建自定义镜像。

#### 2. 修改后端terminal服务

**文件**: `internal/service/terminal.go`

核心改动：

```go
// CreateSession creates or attaches to a persistent tmux session
func (s *TerminalService) CreateSession(ctx context.Context, ws *websocket.Conn, containerID string) error {
    sessionName := fmt.Sprintf("vibox-%s", workspaceID)

    // Check if tmux session exists
    checkCmd := []string{"tmux", "has-session", "-t", sessionName}
    _, err := s.dockerSvc.ExecCommand(ctx, containerID, checkCmd)

    var execCmd []string
    if err != nil {
        // Session doesn't exist - create new one
        utils.Info("Creating new tmux session", "session", sessionName)
        execCmd = []string{
            "tmux", "new-session", "-s", sessionName,
            "-x", "80", "-y", "24", // Initial size
            "bash", // Start bash in the session
        }
    } else {
        // Session exists - attach to it
        utils.Info("Attaching to existing tmux session", "session", sessionName)
        execCmd = []string{
            "tmux", "attach-session", "-t", sessionName,
        }
    }

    // Execute tmux command and attach streams...
    // (类似现有的exec逻辑，但使用tmux命令)
}
```

#### 3. 处理tmux特性

**调整terminal大小**:
```go
func (s *TerminalService) resizeTerminal(ctx context.Context, containerID, sessionName string, cols, rows int) error {
    cmd := []string{
        "tmux", "resize-window",
        "-t", sessionName,
        "-x", strconv.Itoa(cols),
        "-y", strconv.Itoa(rows),
    }
    return s.dockerSvc.ExecCommand(ctx, containerID, cmd)
}
```

**处理多客户端attach**:
- tmux天然支持多个客户端attach到同一会话
- 所有客户端看到相同的terminal内容
- 任何客户端的输入对所有人可见

#### 4. 前端无需改动

前端继续使用xterm.js + WebSocket，对tmux实现透明。

---

### 方案优势

✅ **真正的持久化**: terminal会话在容器内运行，不依赖WebSocket连接
✅ **跨设备访问**: 任何设备登录都能看到同一个terminal
✅ **命令历史保留**: bash历史完整保存
✅ **协作功能**: 多人可以同时操作同一个terminal（如果需要）
✅ **断线重连**: 网络断开后重连，terminal状态完整恢复

### 方案限制

⚠️ **tmux依赖**: 需要容器内安装tmux
⚠️ **内存占用**: tmux会话会占用额外内存
⚠️ **会话清理**: 需要在workspace删除时清理tmux会话

---

## 实现优先级建议

### 已完成 ✅
1. 方向键导航（bash支持）
2. 同浏览器页面切换状态保持

### 下一步（如果需要跨浏览器持久化）

**Option A: 完整tmux方案**（推荐，功能最强）
- 时间估计：2-3天
- 需要修改：
  - `internal/service/terminal.go` （主要修改）
  - 工作空间镜像（预装tmux）
  - 清理逻辑（删除workspace时杀掉tmux session）

**Option B: 轻量级buffer方案**（折中）
- 在后端保存terminal输出历史（最近N行）
- 新连接时回放历史
- 时间估计：1-2天
- 限制：只能回放输出，无法恢复shell状态

**Option C: 保持当前方案**（最简单）
- 只支持同浏览器页面切换
- 刷新页面或换浏览器需要重新连接
- 适合大多数场景

---

## 测试当前修复

服务已重启，你可以：

1. **测试方向键**:
   - 创建新的ubuntu workspace
   - 进入terminal
   - 执行几个命令（如 `ls`, `pwd`, `echo hello`）
   - 按上箭头，应该能看到命令历史

2. **测试页面切换**:
   - 进入terminal，执行一些命令
   - 切换到Workspaces列表页
   - 再切回terminal
   - 应该看到之前的内容还在

3. **测试刷新页面**:
   - 刷新浏览器
   - Terminal内容会清空（这是当前限制）
   - 如果需要保留，需要实现上面的tmux方案

---

## 下一步行动

**请告诉我**:
1. 你是否需要跨浏览器terminal持久化？
2. 如果需要，偏好哪个实现方案（A/B/C）？
3. 还是说当前的修复（方向键+页面切换）已经足够？

我可以：
- 继续实现tmux方案（如果需要）
- 测试和优化当前方案
- 生成完整的测试报告

**当前服务状态**:
- ViBox运行在: http://localhost:3000
- Token: `78215851816230bf3627930215e6c440a7d14ea2f865155b9492d06f16117feb`
- 已修复: 方向键 + 页面切换保留
