# 第二阶段：前端开发

> **当前阶段目标**：实现完整的前端界面，提供直观的工作空间管理体验

---

## 目录

1. [技术栈](#技术栈)
2. [UI/UX 设计理念](#uiux-设计理念)
3. [项目架构](#项目架构)
4. [UI 原型设计](#ui-原型设计)
5. [开发计划](#开发计划)
6. [与后端集成](#与后端集成)

---

## 技术栈

### 核心依赖

```json
{
  "dependencies": {
    "react": "^18.3.1",
    "react-dom": "^18.3.1",
    "react-router-dom": "^6.28.0",
    "@xterm/xterm": "^5.5.0",
    "@xterm/addon-fit": "^0.10.0",
    "@xterm/addon-web-links": "^0.11.0",
    "@xterm/addon-webgl": "^0.18.0",
    "axios": "^1.7.9",
    "jotai": "^2.10.6"
  },
  "devDependencies": {
    "@vitejs/plugin-react": "^4.3.4",
    "vite": "^6.0.7",
    "typescript": "^5.7.3",
    "tailwindcss": "^4.0.0",
    "autoprefixer": "^10.4.20",
    "postcss": "^8.5.1"
  }
}
```

### UI 组件库

- **shadcn UI**：基于 Radix UI 的高质量组件库
- 按需安装，完全可定制
- 完美集成 Tailwind CSS

### 为什么选择这些技术？

- **React 18**：最流行的前端框架，生态系统完善，WebSocket 和终端集成支持好
- **Vite**：超快的开发服务器和构建工具，现代化的开发体验
- **TypeScript**：类型安全，提高代码质量，减少运行时错误
- **Tailwind CSS**：Utility-first CSS 框架，快速开发，生产体积小
- **shadcn UI**：现代美观的组件，完全可控，与 Tailwind 原生集成
- **Jotai**：轻量级原子化状态管理，简单易用，TypeScript 友好
- **xterm.js**：成熟的终端模拟器，支持完整的 ANSI 转义序列

---

## UI/UX 设计理念

### 核心原则

#### 1. 简洁直观
- **一目了然**：工作空间状态（运行中、停止、错误）使用颜色和图标清晰标识
- **操作便捷**：常用操作（打开终端、访问端口）一键直达
- **减少层级**：避免过深的菜单嵌套，扁平化设计

#### 2. 实时反馈
- **状态同步**：工作空间状态实时更新（使用轮询或 WebSocket）
- **操作确认**：创建、删除等操作提供清晰的成功/失败提示
- **加载状态**：异步操作显示加载动画，避免用户疑惑

#### 3. 响应式设计
- **桌面优先**：主要面向桌面用户（开发者）
- **适配平板**：支持 iPad 等平板设备
- **移动端友好**：基础功能在手机上可用

#### 4. 终端体验
- **全屏沉浸**：终端可全屏或最大化，专注于命令行操作
- **快捷键支持**：常用操作支持键盘快捷键
- **主题支持**：终端主题可配置（暗色/亮色）

---

## 项目架构

### 目录结构

```
frontend/
├── src/
│   ├── components/
│   │   ├── ui/                 # shadcn UI 组件
│   │   │   ├── button.tsx
│   │   │   ├── card.tsx
│   │   │   ├── dialog.tsx
│   │   │   ├── badge.tsx
│   │   │   ├── table.tsx
│   │   │   └── ...
│   │   ├── layout/             # 布局组件
│   │   │   ├── Header.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── Layout.tsx
│   │   ├── workspace/          # 工作空间相关组件
│   │   │   ├── WorkspaceCard.tsx
│   │   │   ├── WorkspaceList.tsx
│   │   │   ├── CreateWorkspaceDialog.tsx
│   │   │   └── DeleteConfirmDialog.tsx
│   │   └── terminal/           # 终端组件
│   │       ├── Terminal.tsx
│   │       └── TerminalToolbar.tsx
│   ├── pages/                  # 页面组件
│   │   ├── LoginPage.tsx
│   │   ├── WorkspacesPage.tsx
│   │   ├── WorkspaceDetailPage.tsx
│   │   └── SettingsPage.tsx
│   ├── hooks/                  # 自定义 Hooks
│   │   ├── useWebSocket.ts
│   │   ├── useWorkspaces.ts
│   │   ├── useTerminal.ts
│   │   └── useAuth.ts
│   ├── api/                    # API 调用
│   │   ├── client.ts           # Axios 实例配置
│   │   ├── workspaces.ts       # 工作空间 API
│   │   └── types.ts            # API 类型定义
│   ├── stores/                 # Jotai 状态管理
│   │   ├── auth.ts             # 认证状态
│   │   ├── workspaces.ts       # 工作空间状态
│   │   └── ui.ts               # UI 状态（主题、侧边栏等）
│   ├── lib/                    # 工具函数
│   │   ├── utils.ts            # 通用工具
│   │   └── cn.ts               # className 合并工具
│   ├── types/                  # TypeScript 类型定义
│   │   └── workspace.ts
│   ├── App.tsx                 # 根组件
│   ├── main.tsx                # 入口文件
│   └── index.css               # 全局样式
├── public/
│   └── favicon.ico
├── index.html
├── vite.config.ts
├── tailwind.config.js
├── tsconfig.json
├── postcss.config.js
├── components.json             # shadcn UI 配置
└── package.json
```

### 核心模块

#### 1. 状态管理（Jotai）

Jotai 是一个原子化状态管理库，简单轻量，TypeScript 友好。

```typescript
// stores/auth.ts
import { atom } from 'jotai'

export const tokenAtom = atom<string | null>(
  localStorage.getItem('api_token')
)

export const isAuthenticatedAtom = atom(
  (get) => get(tokenAtom) !== null
)
```

```typescript
// stores/workspaces.ts
import { atom } from 'jotai'
import type { Workspace } from '@/types/workspace'

export const workspacesAtom = atom<Workspace[]>([])

export const selectedWorkspaceIdAtom = atom<string | null>(null)

export const selectedWorkspaceAtom = atom((get) => {
  const workspaces = get(workspacesAtom)
  const selectedId = get(selectedWorkspaceIdAtom)
  return workspaces.find(ws => ws.id === selectedId)
})
```

**为什么选择 Jotai 而不是 Redux/Zustand？**

| 特性 | Jotai | Redux | Zustand |
|------|-------|-------|---------|
| **学习曲线** | ✅ 简单 | ❌ 复杂 | ✅ 简单 |
| **TypeScript 支持** | ✅ 原生 | ⚠️ 需要配置 | ✅ 良好 |
| **代码量** | ✅ 极少 | ❌ 较多 | ✅ 少 |
| **性能** | ✅ 优秀 | ✅ 优秀 | ✅ 优秀 |
| **原子化** | ✅ 原生支持 | ❌ 需要额外库 | ❌ 不支持 |
| **适合场景** | 中小型项目 | 大型复杂项目 | 中型项目 |

**Jotai 的优势**：
- 原子化设计，组件只订阅需要的状态，避免不必要的重渲染
- 代码简洁，几乎无模板代码
- 完美的 TypeScript 类型推导
- 类似 React Hooks 的 API，学习成本低

#### 2. API 客户端（Axios）

```typescript
// api/client.ts
import axios from 'axios'
import { tokenAtom } from '@/stores/auth'
import { getDefaultStore } from 'jotai'

const store = getDefaultStore()

const client = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

// 请求拦截器：自动添加 token
client.interceptors.request.use((config) => {
  const token = store.get(tokenAtom)
  if (token) {
    config.headers['X-ViBox-Token'] = token
  }
  return config
})

// 响应拦截器：统一错误处理
client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token 失效，跳转登录
      store.set(tokenAtom, null)
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default client
```

```typescript
// api/workspaces.ts
import client from './client'
import type { Workspace, CreateWorkspaceRequest, UpdatePortsRequest } from './types'

export const workspaceApi = {
  // 获取所有工作空间
  list: () => client.get<Workspace[]>('/workspaces'),

  // 获取单个工作空间
  get: (id: string) => client.get<Workspace>(`/workspaces/${id}`),

  // 创建工作空间
  create: (data: CreateWorkspaceRequest) =>
    client.post<Workspace>('/workspaces', data),

  // 删除工作空间
  delete: (id: string) => client.delete(`/workspaces/${id}`),

  // 更新端口映射（新增）
  updatePorts: (id: string, data: UpdatePortsRequest) =>
    client.put<Workspace>(`/workspaces/${id}/ports`, data),

  // 重置工作空间（新增）
  reset: (id: string) =>
    client.post<{ message: string; workspace: Workspace }>(`/workspaces/${id}/reset`),
}
```

#### 3. WebSocket 管理

```typescript
// hooks/useWebSocket.ts
import { useEffect, useRef, useState } from 'react'
import { useAtomValue } from 'jotai'
import { tokenAtom } from '@/stores/auth'

export const useWebSocket = (url: string) => {
  const token = useAtomValue(tokenAtom)
  const wsRef = useRef<WebSocket | null>(null)
  const [status, setStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected')

  useEffect(() => {
    if (!token) return

    const wsUrl = `${url}?token=${token}`
    const connect = () => {
      setStatus('connecting')
      const ws = new WebSocket(wsUrl)

      ws.onopen = () => setStatus('connected')
      ws.onclose = () => {
        setStatus('disconnected')
        // 3 秒后重连
        setTimeout(connect, 3000)
      }

      wsRef.current = ws
    }

    connect()

    return () => {
      wsRef.current?.close()
    }
  }, [url, token])

  return { ws: wsRef.current, status }
}
```

#### 4. 终端集成

```typescript
// components/terminal/Terminal.tsx
import { useEffect, useRef } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { WebglAddon } from '@xterm/addon-webgl'
import '@xterm/xterm/css/xterm.css'
import { useWebSocket } from '@/hooks/useWebSocket'

interface TerminalProps {
  workspaceId: string
}

export const TerminalComponent = ({ workspaceId }: TerminalProps) => {
  const terminalRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<Terminal | null>(null)
  const { ws, status } = useWebSocket(`ws://localhost:3000/ws/terminal/${workspaceId}`)

  useEffect(() => {
    if (!terminalRef.current) return

    // 创建终端实例
    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
      },
    })

    // 加载插件
    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()
    const webglAddon = new WebglAddon()

    term.loadAddon(fitAddon)
    term.loadAddon(webLinksAddon)
    term.loadAddon(webglAddon)

    term.open(terminalRef.current)
    fitAddon.fit()

    termRef.current = term

    // WebSocket 消息处理
    if (ws && status === 'connected') {
      term.onData((data) => {
        ws.send(JSON.stringify({ type: 'input', data }))
      })

      ws.onmessage = (event) => {
        const msg = JSON.parse(event.data)
        if (msg.type === 'output') {
          term.write(msg.data)
        }
      }

      // 窗口大小调整
      const handleResize = () => {
        fitAddon.fit()
        ws.send(JSON.stringify({
          type: 'resize',
          cols: term.cols,
          rows: term.rows,
        }))
      }

      window.addEventListener('resize', handleResize)
      return () => window.removeEventListener('resize', handleResize)
    }

    return () => {
      term.dispose()
    }
  }, [workspaceId, ws, status])

  return (
    <div className="h-full w-full bg-[#1e1e1e]">
      <div ref={terminalRef} className="h-full" />
    </div>
  )
}
```

---

## UI 原型设计

### 页面结构

ViBox 前端包含以下核心页面：

```
应用结构
├── 登录页（/login）
│   └── Token 输入
├── 工作空间列表页（/）
│   ├── 顶部导航栏
│   ├── 创建工作空间按钮
│   └── 工作空间卡片列表
├── 工作空间详情页（/workspace/:id）
│   ├── 标签页导航
│   ├── 终端标签页
│   ├── 端口转发标签页
│   └── 配置标签页
└── 设置页（/settings）
    └── Token 管理
```

### 1. 登录页（Login Page）

**目的**：用户输入 API Token 进行认证

**布局**：
```
┌────────────────────────────────────────┐
│                                        │
│          ViBox Logo                    │
│                                        │
│    ┌──────────────────────────┐       │
│    │  API Token               │       │
│    │  [_________________]     │       │
│    │                          │       │
│    │  [ Login Button ]        │       │
│    └──────────────────────────┘       │
│                                        │
│  Token 示例：                          │
│  openssl rand -hex 32                 │
│                                        │
└────────────────────────────────────────┘
```

**组件**：
- shadcn UI `Card` - 登录表单容器
- shadcn UI `Input` - Token 输入框
- shadcn UI `Button` - 登录按钮

**交互**：
1. 用户输入 token
2. 点击登录
3. 验证 token（调用 `/api/workspaces` 测试）
4. 成功 → 跳转工作空间列表
5. 失败 → 显示错误提示

---

### 2. 工作空间列表页（Workspaces Page）

**目的**：展示所有工作空间，提供创建、删除等操作

**布局**：
```
┌─────────────────────────────────────────────────┐
│  ViBox        Workspaces    Settings    Logout  │  ← Header
├─────────────────────────────────────────────────┤
│                                                 │
│  Workspaces                  [+ New Workspace] │
│                                                 │
│  ┌──────────────┐  ┌──────────────┐           │
│  │ dev-env      │  │ test-env     │           │
│  │ ●Running     │  │ ●Running     │           │
│  │              │  │              │           │
│  │ ubuntu:22.04 │  │ alpine       │           │
│  │              │  │              │           │
│  │ Quick Access:│  │ Quick Access:│           │
│  │ [VSCode:8080]│  │ [App:3000]   │           │
│  │ [App:3000]   │  │              │           │
│  │              │  │              │           │
│  │ [Terminal]   │  │ [Terminal]   │           │
│  │ [Ports]      │  │ [Ports]      │           │
│  │ [Delete]     │  │ [Delete]     │           │
│  └──────────────┘  └──────────────┘           │
│                                                 │
│  ┌──────────────┐                              │
│  │ build-env    │                              │
│  │ ⊗Error       │                              │
│  │              │                              │
│  │ node:20      │                              │
│  │              │                              │
│  │ [View Logs]  │                              │
│  │ [Delete]     │                              │
│  └──────────────┘                              │
│                                                 │
└─────────────────────────────────────────────────┘
```

**组件**：
- shadcn UI `Card` - 工作空间卡片
- shadcn UI `Badge` - 状态标识（Running/Stopped/Error）
- shadcn UI `Button` - 操作按钮
- shadcn UI `Dialog` - 创建/删除确认对话框

**工作空间卡片**：
- **顶部**：工作空间名称 + 状态 Badge
- **中部**：镜像信息、端口快速访问按钮
- **底部**：操作按钮（打开终端、查看端口、删除）

**端口快速访问**：
- 显示前 2-3 个端口标签（如果有配置）
- 点击按钮在新窗口打开对应的代理 URL
- 例如：`[VSCode:8080]` 打开 `/forward/{workspace-id}/8080/`
- 如果没有配置端口标签，不显示此区域

**状态颜色**：
- `running` - 绿色（成功）
- `stopped` - 灰色（次要）
- `creating` - 蓝色（信息）
- `error` - 红色（危险）

**交互**：
1. **创建工作空间**：
   - 点击 "+ New Workspace" → 打开对话框
   - 填写表单（名称、镜像、脚本）→ 提交
   - 显示加载状态 → 成功/失败提示
   - 新工作空间出现在列表中

2. **打开终端**：
   - 点击 "Terminal" 按钮
   - 跳转到工作空间详情页的终端标签

3. **查看端口**：
   - 点击 "Ports" 按钮
   - 跳转到工作空间详情页的端口标签

4. **删除工作空间**：
   - 点击 "Delete" 按钮 → 确认对话框
   - 确认 → 删除 → 成功提示
   - 卡片从列表中移除

---

### 3. 创建工作空间对话框（Create Workspace Dialog）

**布局**：
```
┌───────────────────────────────────────┐
│  Create New Workspace            [X]  │
├───────────────────────────────────────┤
│                                       │
│  Workspace Name *                     │
│  [_____________________________]      │
│                                       │
│  Docker Image                         │
│  [ubuntu:22.04          ▼]            │
│                                       │
│  ┌─ Initialization Scripts ────────┐ │
│  │                                  │ │
│  │  Script 1                        │ │
│  │  [install-tools          ▼]     │ │
│  │  ┌────────────────────────────┐ │ │
│  │  │#!/bin/bash                 │ │ │
│  │  │apt-get update             │ │ │
│  │  │apt-get install -y curl... │ │ │
│  │  └────────────────────────────┘ │ │
│  │                                  │ │
│  │  [+ Add Script]                 │ │
│  └──────────────────────────────────┘ │
│                                       │
│  ┌─ Port Labels (Optional) ────────┐ │
│  │                                  │ │
│  │  Port    Service Name            │ │
│  │  [8080]  [VS Code Server____]  ✕ │ │
│  │  [3000]  [Web App___________]  ✕ │ │
│  │                                  │ │
│  │  [+ Add Port]                   │ │
│  └──────────────────────────────────┘ │
│                                       │
│        [Cancel]    [Create]           │
└───────────────────────────────────────┘
```

**组件**：
- shadcn UI `Dialog` - 对话框容器
- shadcn UI `Input` - 工作空间名称、端口号、服务名称
- shadcn UI `Select` - 镜像选择
- shadcn UI `Textarea` - 脚本内容
- shadcn UI `Button` - 添加/删除端口标签
- shadcn UI `Button` - 添加脚本、取消、创建

**表单字段**：
1. **Workspace Name**（必填）
   - 验证：非空、唯一
2. **Docker Image**（可选，默认 ubuntu:22.04）
   - 常用镜像下拉：ubuntu:22.04, alpine:latest, node:20, python:3.11
3. **Initialization Scripts**（可选）
   - 可添加多个脚本
   - 每个脚本有名称、内容、执行顺序
4. **Port Labels**（可选）
   - 为常用端口设置友好的服务名称
   - 每个端口标签包含：端口号（1-65535）、服务名称
   - 用户可添加/删除多个端口标签
   - 示例：8080 → "VS Code Server", 3000 → "Web App"

**交互**：
1. 填写表单
2. 点击 "Create" → 提交 API
3. 显示加载状态
4. 成功 → 关闭对话框，返回列表，显示 Toast 提示
5. 失败 → 显示错误信息

---

### 4. 工作空间详情页（Workspace Detail Page）

**目的**：提供终端访问、端口转发、配置查看

**布局**：
```
┌─────────────────────────────────────────────────┐
│  ← Back to Workspaces                           │
├─────────────────────────────────────────────────┤
│  dev-env                             ●Running   │
│                                                 │
│  ┌─────────────────────────────────────────┐   │
│  │  Terminal  │  Ports  │  Config          │   │  ← Tabs
│  ├─────────────────────────────────────────┤   │
│  │                                          │   │
│  │  Terminal Content                       │   │
│  │  $ _                                     │   │
│  │                                          │   │
│  │                                          │   │
│  │                                          │   │
│  │                                          │   │
│  └─────────────────────────────────────────┘   │
│                                                 │
└─────────────────────────────────────────────────┘
```

**标签页**：

#### 4.1 终端标签页（Terminal Tab）

```
┌─────────────────────────────────────────────┐
│  Terminal                            [🔄] [⛶] │  ← 工具栏
├─────────────────────────────────────────────┤
│  root@container:/# ls -la                   │
│  total 48                                   │
│  drwxr-xr-x  2 root root 4096 Nov 10 12:00 │
│  ...                                        │
│  root@container:/# _                        │
│                                             │
│                                             │
│                                             │
└─────────────────────────────────────────────┘
```

**功能**：
- 🔄 重新连接（WebSocket 断开时）
- ⛶ 全屏/退出全屏
- 显示连接状态（连接中、已连接、已断开）

#### 4.2 端口转发标签页（Ports Tab）

```
┌──────────────────────────────────────────────┐
│  Port Forwarding                     [+ Add] │
├──────────────────────────────────────────────┤
│                                              │
│  Saved Ports                                 │
│                                              │
│  ┌────────────────────────────────────────┐ │
│  │ Service         │ Port │ Actions       │ │
│  ├─────────────────┼──────┼───────────────┤ │
│  │ VS Code Server  │ 8080 │ [Open] [Edit] │ │
│  │ Web App         │ 3000 │ [Open] [Edit] │ │
│  │ PostgreSQL      │ 5432 │ [Open] [Edit] │ │
│  └────────────────────────────────────────┘ │
│                                              │
│  Access Any Port                             │
│  ┌────────────────────────────────────────┐ │
│  │ Port: [____] [Open]                    │ │
│  └────────────────────────────────────────┘ │
│                                              │
│  Note: All ports are accessible dynamically. │
│  Saved ports are for quick access only.     │
│                                              │
└──────────────────────────────────────────────┘
```

**功能**：
- 显示已保存的端口标签（从后端 `ports` 字段读取）
- 点击 [Open] 在新窗口打开端口代理 URL
- 点击 [Edit] 修改服务名称
- [+ Add] 添加新的端口标签
- "Access Any Port" 区域允许手动输入任意端口号访问
- 自动生成代理 URL：`/forward/{workspace-id}/{port}/`

#### 4.3 配置标签页（Config Tab）

```
┌──────────────────────────────────────────────┐
│  Configuration                   [Reset]     │
├──────────────────────────────────────────────┤
│                                              │
│  Workspace ID:     ws-a1b2c3d4              │
│  Name:             dev-env                   │
│  Status:           Running                   │
│  Container ID:     docker-abc123             │
│  Image:            ubuntu:22.04              │
│  Created:          2025-11-10 12:00:00       │
│                                              │
│  ┌─ Initialization Scripts ───────────────┐ │
│  │  1. install-tools                      │ │
│  │     #!/bin/bash                        │ │
│  │     apt-get update && ...              │ │
│  │                                        │ │
│  │  2. setup-user                         │ │
│  │     #!/bin/bash                        │ │
│  │     useradd -m developer               │ │
│  └────────────────────────────────────────┘ │
│                                              │
│  ⚠️ Reset Workspace                          │
│  Deletes the current container and          │
│  recreates it with original configuration.   │
│  All data in the container will be lost.    │
│                                              │
└──────────────────────────────────────────────┘
```

**功能**：
- 只读显示工作空间配置
- 脚本内容展示（代码高亮）
- [Reset] 按钮重置工作空间（需要确认）
  - 删除旧容器
  - 使用原始配置创建新容器
  - 重新执行初始化脚本

---

### 5. 设置页（Settings Page）

**目的**：管理 API Token、应用配置

**布局**：
```
┌─────────────────────────────────────────────────┐
│  ViBox        Workspaces    Settings    Logout  │
├─────────────────────────────────────────────────┤
│                                                 │
│  Settings                                       │
│                                                 │
│  ┌─ API Token ───────────────────────────────┐ │
│  │                                            │ │
│  │  Current Token                             │ │
│  │  [********************]  [Show] [Change]  │ │
│  │                                            │ │
│  │  ⚠️ Keep your token secure!                │ │
│  │     Do not share it with others.           │ │
│  └────────────────────────────────────────────┘ │
│                                                 │
│  ┌─ Theme ───────────────────────────────────┐ │
│  │                                            │ │
│  │  Terminal Theme                            │ │
│  │  ○ Dark (Default)  ○ Light                │ │
│  └────────────────────────────────────────────┘ │
│                                                 │
│  ┌─ About ───────────────────────────────────┐ │
│  │                                            │ │
│  │  ViBox v0.1.0                              │ │
│  │  Backend: Go 1.25                          │ │
│  │  Frontend: React 18 + Vite                │ │
│  └────────────────────────────────────────────┘ │
│                                                 │
└─────────────────────────────────────────────────┘
```

**组件**：
- shadcn UI `Card` - 设置分组
- shadcn UI `Input` - Token 输入
- shadcn UI `RadioGroup` - 主题选择

---

### 设计规范

#### 颜色系统

使用 Tailwind CSS 默认颜色 + shadcn UI 主题：

```css
/* 主色调 */
--primary: 222.2 47.4% 11.2%;      /* 深色 */
--primary-foreground: 210 40% 98%; /* 白色 */

/* 状态颜色 */
--success: 142 76% 36%;    /* 绿色 - Running */
--warning: 48 96% 53%;     /* 黄色 - Creating */
--destructive: 0 84% 60%;  /* 红色 - Error */
--muted: 210 40% 96%;      /* 灰色 - Stopped */
```

#### 字体

```css
font-family:
  /* UI 文本 */
  system-ui, -apple-system, sans-serif

  /* 代码/终端 */
  'Menlo', 'Monaco', 'Courier New', monospace
```

#### 间距

- 卡片间距：`gap-4`（16px）
- 内边距：`p-6`（24px）
- 按钮高度：`h-10`（40px）

#### 圆角

- 卡片：`rounded-lg`（8px）
- 按钮：`rounded-md`（6px）
- 输入框：`rounded-md`（6px）

---

## 开发计划

### 时间规划（9-14 天）

| 阶段 | 任务 | 时间 | 状态 |
|------|------|------|------|
| **Phase 0** | UI 原型设计 | 1-2 天 | 📍 当前阶段 |
| **Phase 1** | 项目初始化 + 基础设施 | 1-2 天 | ⏳ 待开始 |
| **Phase 2** | 认证 + 工作空间列表 | 2-3 天 | ⏳ 待开始 |
| **Phase 3** | 终端集成（WebSSH） | 3-4 天 | ⏳ 待开始 |
| **Phase 4** | 端口转发界面 | 1-2 天 | ⏳ 待开始 |
| **Phase 5** | 优化与测试 | 2-3 天 | ⏳ 待开始 |

---

### Phase 0: UI 原型设计（1-2 天）📍 当前

**目标**：完成界面设计和用户流程规划，为后续开发提供清晰的蓝图

**任务清单**：
- [x] 确定技术栈（React + Vite + Tailwind + shadcn UI + Jotai）
- [x] 设计页面结构和布局
- [x] 定义组件层次
- [x] 规划用户交互流程
- [ ] 创建原型图（可选：使用 Figma/手绘）
- [ ] 评审设计方案
- [ ] 确认设计细节

**交付物**：
- ✅ UI 原型设计文档（本文档）
- ⏳ 原型图（可选）
- ⏳ 设计评审通过

**验收标准**：
- 所有核心页面的布局已定义
- 用户交互流程清晰
- 组件划分合理
- 技术栈选型确认

**下一步**：
完成 UI 原型设计后，将进入 Phase 1（项目初始化），届时会编写详细的任务拆分文档（Task Breakdown）。

---

### Phase 1: 项目初始化 + 基础设施（1-2 天）⏳ 待开始

**任务清单**：
- [ ] 初始化 Vite + React + TypeScript 项目
- [ ] 配置 Tailwind CSS
- [ ] 安装和配置 shadcn UI
- [ ] 配置路由（React Router）
- [ ] 配置 Axios + API 客户端
- [ ] 配置 Jotai 状态管理
- [ ] 创建基础布局组件（Header, Layout）
- [ ] 配置 TypeScript 类型定义

**验收标准**：
- 项目可以启动（`npm run dev`）
- Tailwind CSS 工作正常
- shadcn UI 组件可以导入使用
- 路由配置完成
- API 客户端可以调用后端

---

### Phase 2: 认证 + 工作空间列表（2-3 天）⏳ 待开始

**任务清单**：
- [ ] 实现登录页（Token 输入）
- [ ] 实现认证逻辑（Jotai + localStorage）
- [ ] 实现工作空间列表页
- [ ] 实现工作空间卡片组件
- [ ] 实现创建工作空间对话框
- [ ] 实现删除确认对话框
- [ ] 实现状态轮询（自动刷新列表）
- [ ] 实现 Toast 提示

**验收标准**：
- 可以登录（输入 token）
- 可以查看工作空间列表
- 可以创建工作空间
- 可以删除工作空间
- 状态实时更新

---

### Phase 3: 终端集成（WebSSH）（3-4 天）⏳ 待开始

**任务清单**：
- [ ] 实现工作空间详情页
- [ ] 实现标签页导航
- [ ] 集成 xterm.js
- [ ] 实现 WebSocket 连接
- [ ] 实现终端输入/输出
- [ ] 实现终端 resize
- [ ] 实现重连机制
- [ ] 实现全屏模式
- [ ] 实现连接状态显示

**验收标准**：
- 可以打开终端
- 可以输入命令
- 可以看到输出
- 终端大小自适应
- 连接断开后自动重连

---

### Phase 4: 端口转发界面（1-2 天）⏳ 待开始

**任务清单**：
- [ ] 实现端口转发标签页
- [ ] 实现端口列表展示
- [ ] 实现 URL 复制功能
- [ ] 实现配置标签页
- [ ] 显示工作空间详细信息
- [ ] 显示脚本内容

**验收标准**：
- 可以查看端口转发说明
- 可以复制端口 URL
- 可以查看工作空间配置

---

### Phase 5: 优化与测试（2-3 天）⏳ 待开始

**任务清单**：
- [ ] 响应式布局优化
- [ ] 性能优化（lazy loading, code splitting）
- [ ] 错误处理优化
- [ ] 加载状态优化
- [ ] 无障碍支持（a11y）
- [ ] 浏览器兼容性测试
- [ ] 用户体验优化
- [ ] 编写用户文档

**验收标准**：
- 响应式布局良好
- 加载速度快
- 错误提示友好
- 支持键盘操作
- 主流浏览器兼容

---

## 与后端集成

### 集成方案：Go 后端嵌入前端静态文件

#### 构建流程

```bash
# 1. 构建前端
cd frontend
npm run build
# 输出到 frontend/dist/

# 2. 复制到 Go 后端
cp -r dist ../cmd/server/dist

# 3. Go 嵌入静态文件
# 在 cmd/server/main.go 中使用 embed
```

#### Go 后端配置

```go
// cmd/server/main.go
package main

import (
    "embed"
    "io/fs"
    "net/http"

    "github.com/gin-gonic/gin"
)

//go:embed dist
var staticFiles embed.FS

func main() {
    r := gin.Default()

    // API 路由（需要鉴权）
    api := r.Group("/api", authMiddleware)
    {
        api.GET("/workspaces", listWorkspaces)
        api.POST("/workspaces", createWorkspace)
        // ...
    }

    // WebSocket 路由（需要鉴权）
    r.GET("/ws/terminal/:id", wsAuthMiddleware, handleTerminal)

    // 端口转发路由（需要鉴权）
    r.Any("/forward/:id/:port/*path", proxyAuthMiddleware, handleProxy)

    // 健康检查（无需鉴权）
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // 前端静态文件（无需鉴权）
    staticFS, _ := fs.Sub(staticFiles, "dist")
    r.NoRoute(func(c *gin.Context) {
        // 处理 SPA 路由
        path := c.Request.URL.Path
        if _, err := staticFS.Open(path); err != nil {
            // 文件不存在，返回 index.html（SPA 路由）
            c.FileFromFS("/", http.FS(staticFS))
        } else {
            c.FileFromFS(path, http.FS(staticFS))
        }
    })

    r.Run(":3000")
}
```

#### Vite 配置

```typescript
// frontend/vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      // 开发环境代理 API 到后端
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:3000',
        ws: true,
      },
      '/forward': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          // 代码分割
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'xterm-vendor': ['@xterm/xterm', '@xterm/addon-fit', '@xterm/addon-web-links'],
        },
      },
    },
  },
})
```

#### 开发环境

```bash
# 终端 1: 启动 Go 后端
cd vibox
go run ./cmd/server

# 终端 2: 启动前端开发服务器
cd frontend
npm run dev

# 访问 http://localhost:5173
# Vite 会自动代理 API 请求到 :3000
```

#### 生产环境

```bash
# 构建前端
cd frontend
npm run build

# 复制静态文件到 Go 项目
cp -r dist ../cmd/server/

# 构建 Go 后端（包含嵌入的前端）
cd ..
go build -o vibox ./cmd/server

# 运行单一可执行文件
./vibox

# 访问 http://localhost:3000
```

#### Docker 集成

```dockerfile
# Dockerfile
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/dist ./cmd/server/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o vibox ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=backend-builder /app/vibox .
EXPOSE 3000
CMD ["./vibox"]
```

---

## API 集成示例

### 类型定义

```typescript
// src/types/workspace.ts
export interface Workspace {
  id: string
  name: string
  container_id: string
  status: 'creating' | 'running' | 'stopped' | 'error'
  created_at: string
  config: WorkspaceConfig
  ports?: Record<string, string>  // 新增：端口标签映射
  error?: string
}

export interface WorkspaceConfig {
  image: string
  scripts: Script[]
}

export interface Script {
  name: string
  content: string
  order: number
}

export interface CreateWorkspaceRequest {
  name: string
  image?: string
  scripts?: Script[]
  ports?: Record<string, string>  // 新增：端口标签映射
}

export interface UpdatePortsRequest {
  ports: Record<string, string>
}
```

### API 调用示例

```typescript
// src/hooks/useWorkspaces.ts
import { useAtom } from 'jotai'
import { useEffect } from 'react'
import { workspacesAtom } from '@/stores/workspaces'
import { workspaceApi } from '@/api/workspaces'

export const useWorkspaces = () => {
  const [workspaces, setWorkspaces] = useAtom(workspacesAtom)

  const fetchWorkspaces = async () => {
    try {
      const { data } = await workspaceApi.list()
      setWorkspaces(data)
    } catch (error) {
      console.error('Failed to fetch workspaces:', error)
    }
  }

  useEffect(() => {
    fetchWorkspaces()

    // 每 5 秒轮询一次
    const interval = setInterval(fetchWorkspaces, 5000)
    return () => clearInterval(interval)
  }, [])

  return { workspaces, refetch: fetchWorkspaces }
}

// 更新端口标签示例
export const updateWorkspacePorts = async (
  workspaceId: string,
  ports: Record<string, string>
) => {
  try {
    const { data } = await workspaceApi.updatePorts(workspaceId, { ports })
    console.log('Ports updated successfully:', data)
    return data
  } catch (error) {
    console.error('Failed to update ports:', error)
    throw error
  }
}

// 重置工作空间示例
export const resetWorkspace = async (workspaceId: string) => {
  try {
    const { data } = await workspaceApi.reset(workspaceId)
    console.log('Workspace reset successfully:', data.message)
    return data.workspace
  } catch (error) {
    console.error('Failed to reset workspace:', error)
    throw error
  }
}

// 使用示例：在组件中
const handleUpdatePorts = async () => {
  await updateWorkspacePorts('ws-12345', {
    '8080': 'VS Code Server',
    '3000': 'Web App',
    '5432': 'PostgreSQL'
  })
  refetch() // 重新获取工作空间列表
}

const handleReset = async () => {
  if (confirm('Are you sure? This will delete the container and recreate it.')) {
    await resetWorkspace('ws-12345')
    refetch()
  }
}
```

---

## 成功标准

第二阶段完成后，应该能够：

- ✅ 通过 Web 界面输入 Token 登录
- ✅ 查看所有工作空间列表
- ✅ 创建新的工作空间（配置镜像和脚本）
- ✅ 打开工作空间的 Web 终端
- ✅ 在终端中执行命令
- ✅ 查看端口转发说明并复制 URL
- ✅ 删除工作空间
- ✅ 响应式布局，支持桌面和平板
- ✅ 前端构建产物嵌入 Go 后端，单一可执行文件部署

**下一步**：进入[第三阶段](../PROJECT_ROADMAP.md#第三阶段完整功能扩展--待定)，实现 GitHub 集成、AI Agent、VS Code Server 等高级功能。

---

## 参考资源

### 技术文档
- [React 官方文档](https://react.dev/)
- [Vite 官方文档](https://vitejs.dev/)
- [Tailwind CSS 文档](https://tailwindcss.com/)
- [shadcn UI 文档](https://ui.shadcn.com/)
- [Jotai 文档](https://jotai.org/)
- [xterm.js 文档](https://xtermjs.org/)

### 设计参考
- [Vercel Dashboard](https://vercel.com/) - 简洁的工作空间管理界面
- [Render Dashboard](https://render.com/) - 清晰的状态展示
- [Railway App](https://railway.app/) - 现代的开发者工具界面

---

**版本**: v1.0.0
**日期**: 2025-11-10
**状态**: 📍 Phase 0 - UI 原型设计中
