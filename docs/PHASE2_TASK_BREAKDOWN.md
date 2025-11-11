# Phase 2 Frontend Development Task Breakdown

> **Goal**: Break down frontend development into independent modules to support parallel development by multiple agents

---

## Table of Contents

1. [Module Dependency Graph](#module-dependency-graph)
2. [Parallel Development Strategy](#parallel-development-strategy)
3. [Detailed Module Descriptions](#detailed-module-descriptions)
4. [Interface Definitions](#interface-definitions)
5. [Development Schedule](#development-schedule)
6. [Team Collaboration Guidelines](#team-collaboration-guidelines)

---

## Module Dependency Graph

```
┌─────────────────────────────────────────────────────────┐
│  Layer 1: Foundation (基础设施层)                        │
│  - Project Setup (Vite + React + TS)                   │
│  - Tailwind CSS + shadcn UI                            │
│  - Router Configuration                                │
│  - API Client (Axios)                                  │
└─────────────────┬───────────────────────────────────────┘
                  │
    ┌─────────────┴─────────────┐
    │                           │
┌───▼────────────────┐   ┌─────▼─────────────────────────┐
│  Layer 2: State    │   │  Layer 2: API Integration     │
│  - Jotai Atoms     │   │  - Auth API                   │
│  - Auth State      │   │  - Workspace API              │
│  - Workspace State │   │  - Type Definitions           │
└───┬────────────────┘   └─────┬─────────────────────────┘
    │                           │
    │    ┌──────────────────────┴────────────────────────┐
    │    │                      │                        │
┌───▼────▼──────────┐  ┌───────▼───────┐  ┌────────────▼─────────┐
│  Layer 3: Core UI  │  │  Layer 3:     │  │  Layer 3:            │
│  - Auth Pages      │  │  Terminal     │  │  Workspace Pages     │
│  - Layout          │  │  Integration  │  │  - List              │
│  - shadcn UI       │  │  - xterm.js   │  │  - Detail            │
└───┬────────────────┘  └───────┬───────┘  └────────────┬─────────┘
    │                           │                        │
    └───────────────┬───────────┴────────────────────────┘
                    │
            ┌───────▼────────┐
            │  Layer 4:      │
            │  Integration   │
            │  & Testing     │
            └────────────────┘
```

**Dependency Notes**:
- **Layer 1** → All other layers depend on it (foundation)
- **Layer 2** → Layer 3, Layer 4 depend on it
- **Layer 3** → Layer 4 depends on it
- **Layer 4** → Final integration and testing

---

## Parallel Development Strategy

### 🚀 Development Rounds

| Round | Modules | Time | Parallel Agents |
|-------|---------|------|-----------------|
| **Round 1** | Module 1: Foundation Layer | 1-2 days | 1-2 agents |
| **Round 2** | Module 2: State Management + Module 3: API Integration | 1-2 days | 2 agents |
| **Round 3** | Module 4: Auth UI + Module 5: Workspace UI + Module 6: Terminal | 3-4 days | 3 agents |
| **Round 4** | Module 7: Integration & Polish | 2-3 days | 2 agents |
| **Round 5** | Module 8: Testing & Optimization | 2-3 days | 1-2 agents |

**Total Estimated Time**: 9-14 days (consistent with original plan, accelerated through parallelization)

---

## Detailed Module Descriptions

### Module 1: Foundation Layer (基础设施层)

**Owner**: Agent 1
**Estimated Time**: 1-2 days
**Priority**: 🔴 Highest (all modules depend on this)

#### 📦 Components

1. **Project Initialization** (`frontend/`)
2. **Tailwind CSS Configuration** (`tailwind.config.js`)
3. **shadcn UI Setup** (`components.json`)
4. **Router Configuration** (`src/App.tsx`)
5. **Base Layout** (`src/components/layout/`)

#### 📋 Task Checklist

- [ ] Initialize Vite + React + TypeScript project
  - [ ] `npm create vite@latest frontend -- --template react-ts`
  - [ ] Configure `tsconfig.json` with path aliases
  - [ ] Set up `vite.config.ts` with proxy configuration
- [ ] Install and configure Tailwind CSS
  - [ ] `npm install -D tailwindcss postcss autoprefixer`
  - [ ] Run `npx tailwindcss init -p`
  - [ ] Configure `tailwind.config.js` with theme
  - [ ] Set up `src/index.css` with Tailwind directives
- [ ] Install and configure shadcn UI
  - [ ] Run `npx shadcn-ui@latest init`
  - [ ] Configure `components.json`
  - [ ] Install initial components: button, card, input, dialog, badge
- [ ] Configure React Router
  - [ ] Install `react-router-dom`
  - [ ] Create route structure in `App.tsx`
  - [ ] Set up route guards (authentication check)
- [ ] Create base layout components
  - [ ] `Layout.tsx` - Main layout wrapper
  - [ ] `Header.tsx` - Top navigation bar
  - [ ] `Sidebar.tsx` - Side navigation (for workspace detail)
- [ ] Set up development environment
  - [ ] Configure ESLint and Prettier
  - [ ] Set up `.env` file for environment variables
  - [ ] Configure Vite proxy for backend API

#### 🔌 External Interfaces

**Vite Configuration**:
```typescript
// vite.config.ts
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
})
```

**Router Structure**:
```typescript
// src/App.tsx
const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  {
    path: '/',
    element: <ProtectedRoute><Layout /></ProtectedRoute>,
    children: [
      { index: true, element: <WorkspacesPage /> },
      { path: 'workspace/:id', element: <WorkspaceDetailPage /> },
      { path: 'settings', element: <SettingsPage /> },
    ],
  },
])
```

**Layout Components**:
```typescript
// src/components/layout/Layout.tsx
export const Layout = () => {
  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="container mx-auto py-6">
        <Outlet />
      </main>
    </div>
  )
}
```

#### ✅ Acceptance Criteria

- [ ] Project starts successfully (`npm run dev`)
- [ ] Tailwind CSS classes work correctly
- [ ] shadcn UI components can be imported and used
- [ ] Routes navigate correctly
- [ ] API proxy works (test with `/api/health` if available)
- [ ] TypeScript compilation has no errors
- [ ] Development hot reload works

#### 📚 Dependencies

- Node.js 18+
- npm or pnpm
- `vite`, `react`, `react-dom`, `react-router-dom`
- `tailwindcss`, `autoprefixer`, `postcss`
- shadcn UI components

---

### Module 2: State Management (状态管理)

**Owner**: Agent 2
**Estimated Time**: 1 day
**Priority**: 🟡 Medium (can parallel with Module 3)
**Dependencies**: Module 1

#### 📦 Components

1. **Jotai Atoms** (`src/stores/`)
   - `auth.ts` - Authentication state
   - `workspaces.ts` - Workspace state
   - `ui.ts` - UI state (theme, sidebar, etc.)

#### 📋 Task Checklist

- [ ] Install Jotai
  - [ ] `npm install jotai`
- [ ] Define authentication atoms
  - [ ] `tokenAtom` - API token (sync with localStorage)
  - [ ] `isAuthenticatedAtom` - Derived authentication state
- [ ] Define workspace atoms
  - [ ] `workspacesAtom` - Workspace list
  - [ ] `selectedWorkspaceIdAtom` - Currently selected workspace
  - [ ] `selectedWorkspaceAtom` - Derived selected workspace
- [ ] Define UI atoms
  - [ ] `themeAtom` - Terminal theme (dark/light)
  - [ ] `sidebarOpenAtom` - Sidebar open/close state
- [ ] Create custom hooks
  - [ ] `useAuth()` - Authentication utilities
  - [ ] `useWorkspaces()` - Workspace management utilities
- [ ] Set up localStorage persistence
  - [ ] Sync `tokenAtom` with localStorage
  - [ ] Sync `themeAtom` with localStorage

#### 🔌 External Interfaces

**Auth State**:
```typescript
// src/stores/auth.ts
import { atom } from 'jotai'

export const tokenAtom = atom<string | null>(
  localStorage.getItem('api_token')
)

export const isAuthenticatedAtom = atom(
  (get) => get(tokenAtom) !== null
)

// Writable atom with side effect
export const setTokenAtom = atom(
  null,
  (get, set, newToken: string | null) => {
    set(tokenAtom, newToken)
    if (newToken) {
      localStorage.setItem('api_token', newToken)
    } else {
      localStorage.removeItem('api_token')
    }
  }
)
```

**Workspace State**:
```typescript
// src/stores/workspaces.ts
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

**Custom Hooks**:
```typescript
// src/hooks/useAuth.ts
import { useAtom } from 'jotai'
import { tokenAtom, setTokenAtom, isAuthenticatedAtom } from '@/stores/auth'

export const useAuth = () => {
  const [token] = useAtom(tokenAtom)
  const [isAuthenticated] = useAtom(isAuthenticatedAtom)
  const [, setToken] = useAtom(setTokenAtom)

  const logout = () => {
    setToken(null)
  }

  return { token, isAuthenticated, setToken, logout }
}
```

#### ✅ Acceptance Criteria

- [ ] Atoms are defined and properly typed
- [ ] localStorage sync works for token
- [ ] Custom hooks work correctly
- [ ] State updates trigger re-renders
- [ ] Derived atoms compute correctly
- [ ] No memory leaks or stale closures

#### 📚 Dependencies

- Module 1
- `jotai`

---

### Module 3: API Integration (API集成)

**Owner**: Agent 2 (continue) or Agent 3
**Estimated Time**: 1 day
**Priority**: 🟡 Medium (can parallel with Module 2)
**Dependencies**: Module 1

#### 📦 Components

1. **API Client** (`src/api/client.ts`)
2. **Auth API** (`src/api/auth.ts`)
3. **Workspace API** (`src/api/workspaces.ts`)
4. **Type Definitions** (`src/api/types.ts`, `src/types/`)

#### 📋 Task Checklist

- [ ] Install Axios
  - [ ] `npm install axios`
- [ ] Create Axios client instance
  - [ ] Configure base URL
  - [ ] Set `withCredentials: true` for Cookie support
  - [ ] Add response interceptor for 401 errors
- [ ] Define TypeScript types
  - [ ] `Workspace` - Workspace model
  - [ ] `WorkspaceConfig` - Configuration model
  - [ ] `Script` - Script model
  - [ ] `CreateWorkspaceRequest` - API request types
  - [ ] `UpdatePortsRequest` - Port update request
- [ ] Implement Auth API
  - [ ] `login(token: string)` - Login and set Cookie
  - [ ] `logout()` - Logout and clear Cookie
- [ ] Implement Workspace API
  - [ ] `list()` - Get all workspaces
  - [ ] `get(id: string)` - Get workspace by ID
  - [ ] `create(data: CreateWorkspaceRequest)` - Create workspace
  - [ ] `delete(id: string)` - Delete workspace
  - [ ] `updatePorts(id: string, data: UpdatePortsRequest)` - Update port mappings
  - [ ] `reset(id: string)` - Reset workspace

#### 🔌 External Interfaces

**Axios Client**:
```typescript
// src/api/client.ts
import axios from 'axios'
import { getDefaultStore } from 'jotai'
import { setTokenAtom } from '@/stores/auth'

const store = getDefaultStore()

const client = axios.create({
  baseURL: '/api',
  timeout: 30000,
  withCredentials: true,  // Automatically send Cookie
})

// Response interceptor: handle 401 errors
client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Clear token and redirect to login
      store.set(setTokenAtom, null)
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default client
```

**Auth API**:
```typescript
// src/api/auth.ts
import client from './client'

export const authApi = {
  login: (token: string) =>
    client.post('/auth/login', { token }),
  logout: () =>
    client.post('/auth/logout'),
}
```

**Workspace API**:
```typescript
// src/api/workspaces.ts
import client from './client'
import type { Workspace, CreateWorkspaceRequest, UpdatePortsRequest } from './types'

export const workspaceApi = {
  list: () =>
    client.get<Workspace[]>('/workspaces'),

  get: (id: string) =>
    client.get<Workspace>(`/workspaces/${id}`),

  create: (data: CreateWorkspaceRequest) =>
    client.post<Workspace>('/workspaces', data),

  delete: (id: string) =>
    client.delete(`/workspaces/${id}`),

  updatePorts: (id: string, data: UpdatePortsRequest) =>
    client.put<Workspace>(`/workspaces/${id}/ports`, data),

  reset: (id: string) =>
    client.post<{ message: string; workspace: Workspace }>(`/workspaces/${id}/reset`),
}
```

**Type Definitions**:
```typescript
// src/api/types.ts
export interface Workspace {
  id: string
  name: string
  container_id?: string
  status: 'creating' | 'running' | 'error' | 'failed'
  created_at: string
  config: WorkspaceConfig
  ports?: Record<string, string>
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
  ports?: Record<string, string>
}

export interface UpdatePortsRequest {
  ports: Record<string, string>
}
```

#### ✅ Acceptance Criteria

- [ ] Axios client configured correctly
- [ ] Cookie is automatically sent with requests
- [ ] 401 errors trigger logout and redirect
- [ ] All API methods are typed correctly
- [ ] API methods can be called successfully (mock or real backend)
- [ ] Error handling works properly

#### 📚 Dependencies

- Module 1
- `axios`

---

### Module 4: Authentication UI (认证界面)

**Owner**: Agent 3
**Estimated Time**: 1-2 days
**Priority**: 🟡 Medium
**Dependencies**: Module 1, Module 2, Module 3

#### 📦 Components

1. **Login Page** (`src/pages/LoginPage.tsx`)
2. **Protected Route** (`src/components/ProtectedRoute.tsx`)

#### 📋 Task Checklist

- [ ] Create Login Page component
  - [ ] Token input field
  - [ ] Login button
  - [ ] Loading state during login
  - [ ] Error message display
  - [ ] Example token generation command
- [ ] Implement login logic
  - [ ] Call `authApi.login(token)`
  - [ ] On success: save token to state, redirect to home
  - [ ] On failure: show error message
- [ ] Create ProtectedRoute component
  - [ ] Check authentication state
  - [ ] Redirect to login if not authenticated
  - [ ] Show loading spinner during check
- [ ] Style Login Page
  - [ ] Use shadcn UI Card for form container
  - [ ] Center the form on the page
  - [ ] Add ViBox logo (optional)
  - [ ] Responsive design

#### 🔌 External Interfaces

**Login Page**:
```typescript
// src/pages/LoginPage.tsx
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAtom } from 'jotai'
import { setTokenAtom } from '@/stores/auth'
import { authApi } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'

export const LoginPage = () => {
  const [token, setToken] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [, saveToken] = useAtom(setTokenAtom)
  const navigate = useNavigate()

  const handleLogin = async () => {
    if (!token.trim()) {
      setError('Token is required')
      return
    }

    setLoading(true)
    setError('')

    try {
      await authApi.login(token)
      saveToken(token)
      navigate('/')
    } catch (err) {
      setError('Invalid token')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>ViBox Login</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <Input
              type="password"
              placeholder="Enter API Token"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleLogin()}
            />
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button onClick={handleLogin} disabled={loading} className="w-full">
              {loading ? 'Logging in...' : 'Login'}
            </Button>
            <p className="text-sm text-muted-foreground">
              Generate token: <code>openssl rand -hex 32</code>
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
```

**Protected Route**:
```typescript
// src/components/ProtectedRoute.tsx
import { Navigate } from 'react-router-dom'
import { useAtom } from 'jotai'
import { isAuthenticatedAtom } from '@/stores/auth'

export const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const [isAuthenticated] = useAtom(isAuthenticatedAtom)

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}
```

#### ✅ Acceptance Criteria

- [ ] Login page renders correctly
- [ ] Token input works
- [ ] Login button triggers API call
- [ ] Success: redirects to home page
- [ ] Failure: displays error message
- [ ] ProtectedRoute blocks unauthenticated users
- [ ] Responsive design on mobile/tablet

#### 📚 Dependencies

- Module 1, Module 2, Module 3
- shadcn UI components: Card, Input, Button

---

### Module 5: Workspace UI (工作空间界面)

**Owner**: Agent 4
**Estimated Time**: 2-3 days
**Priority**: 🟡 Medium
**Dependencies**: Module 1, Module 2, Module 3

#### 📦 Components

1. **Workspaces Page** (`src/pages/WorkspacesPage.tsx`)
2. **Workspace Card** (`src/components/workspace/WorkspaceCard.tsx`)
3. **Create Workspace Dialog** (`src/components/workspace/CreateWorkspaceDialog.tsx`)
4. **Delete Confirmation Dialog** (`src/components/workspace/DeleteConfirmDialog.tsx`)
5. **Workspace Detail Page** (`src/pages/WorkspaceDetailPage.tsx`)
6. **Ports Management Tab** (`src/components/workspace/PortsTab.tsx`)
7. **Config Tab** (`src/components/workspace/ConfigTab.tsx`)

#### 📋 Task Checklist

**Part 1: Workspace List**
- [ ] Create WorkspacesPage component
  - [ ] Header with "Create Workspace" button
  - [ ] Grid layout for workspace cards
  - [ ] Loading state
  - [ ] Empty state (no workspaces)
- [ ] Create WorkspaceCard component
  - [ ] Display workspace name and status badge
  - [ ] Display image and container info
  - [ ] Port quick access buttons (if ports configured)
  - [ ] Action buttons: Terminal, Ports, Reset, Delete
  - [ ] Handle error state display (show error message)
  - [ ] Conditional Terminal button (only if container exists)
- [ ] Create CreateWorkspaceDialog component
  - [ ] Form fields: name, image, scripts, ports
  - [ ] Dynamic script list (add/remove)
  - [ ] Dynamic port labels (add/remove)
  - [ ] Form validation
  - [ ] Submit and create workspace
- [ ] Create DeleteConfirmDialog component
  - [ ] Confirmation message
  - [ ] Delete action
- [ ] Implement auto-refresh
  - [ ] Poll workspace list every 5 seconds
  - [ ] Use setInterval + cleanup

**Part 2: Workspace Detail**
- [ ] Create WorkspaceDetailPage component
  - [ ] Tab navigation (Terminal, Ports, Config)
  - [ ] Display workspace name and status in header
  - [ ] Back to list button
- [ ] Create PortsTab component
  - [ ] Display saved port labels table
  - [ ] Edit/Delete port buttons
  - [ ] Add new port button
  - [ ] "Access Any Port" input
  - [ ] Open port in new window
- [ ] Create ConfigTab component
  - [ ] Display workspace ID, name, status
  - [ ] Display container ID and image
  - [ ] Display creation time
  - [ ] Display scripts with syntax highlighting
  - [ ] Reset workspace button

#### 🔌 External Interfaces

**Workspace Card**:
```typescript
// src/components/workspace/WorkspaceCard.tsx
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import type { Workspace } from '@/api/types'

interface WorkspaceCardProps {
  workspace: Workspace
  onDelete: (id: string) => void
  onReset: (id: string) => void
}

export const WorkspaceCard = ({ workspace, onDelete, onReset }: WorkspaceCardProps) => {
  const statusConfig = {
    creating: { color: 'blue', label: 'Creating...', canUseTerminal: false },
    running: { color: 'green', label: 'Running', canUseTerminal: true },
    error: { color: 'orange', label: 'Error', canUseTerminal: !!workspace.container_id },
    failed: { color: 'red', label: 'Failed', canUseTerminal: false },
  }

  const config = statusConfig[workspace.status]

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          {workspace.name}
          <Badge variant={config.color}>{config.label}</Badge>
        </CardTitle>
      </CardHeader>

      <CardContent className="space-y-2">
        <p className="text-sm text-muted-foreground">{workspace.config.image}</p>

        {/* Error message */}
        {workspace.error && (
          <Alert variant="destructive">
            <AlertDescription>{workspace.error}</AlertDescription>
          </Alert>
        )}

        {/* Port quick access */}
        {workspace.status === 'running' && workspace.ports && (
          <div className="flex flex-wrap gap-2">
            {Object.entries(workspace.ports).map(([port, label]) => (
              <Button
                key={port}
                size="sm"
                variant="outline"
                onClick={() => window.open(`/forward/${workspace.id}/${port}/`)}
              >
                {label}:{port}
              </Button>
            ))}
          </div>
        )}
      </CardContent>

      <CardFooter className="flex gap-2">
        <Button disabled={!config.canUseTerminal}>Terminal</Button>
        <Button disabled={workspace.status !== 'running'}>Ports</Button>
        <Button onClick={() => onReset(workspace.id)}>Reset</Button>
        <Button onClick={() => onDelete(workspace.id)} variant="destructive">Delete</Button>
      </CardFooter>
    </Card>
  )
}
```

**Custom Hook for Workspaces**:
```typescript
// src/hooks/useWorkspaces.ts
import { useEffect } from 'react'
import { useAtom } from 'jotai'
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

    // Poll every 5 seconds
    const interval = setInterval(fetchWorkspaces, 5000)
    return () => clearInterval(interval)
  }, [])

  return { workspaces, refetch: fetchWorkspaces }
}
```

#### ✅ Acceptance Criteria

- [ ] Workspace list displays correctly
- [ ] Can create new workspace
- [ ] Can delete workspace
- [ ] Can reset workspace
- [ ] Status updates automatically (polling)
- [ ] Error state displays correctly
- [ ] Port quick access buttons work
- [ ] Terminal button only shows when container exists
- [ ] Workspace detail page renders correctly
- [ ] All tabs work properly
- [ ] Port management works (add/edit/delete)
- [ ] Config tab displays all information

#### 📚 Dependencies

- Module 1, Module 2, Module 3
- shadcn UI components: Card, Badge, Button, Dialog, Table, Alert

---

### Module 6: Terminal Integration (终端集成)

**Owner**: Agent 5
**Estimated Time**: 3-4 days
**Priority**: 🔴 High (core feature)
**Dependencies**: Module 1, Module 2, Module 3

#### 📦 Components

1. **Terminal Component** (`src/components/terminal/Terminal.tsx`)
2. **Terminal Toolbar** (`src/components/terminal/TerminalToolbar.tsx`)
3. **useWebSocket Hook** (`src/hooks/useWebSocket.ts`)
4. **useTerminal Hook** (`src/hooks/useTerminal.ts`)

#### 📋 Task Checklist

- [ ] Install xterm.js and addons
  - [ ] `npm install @xterm/xterm @xterm/addon-fit @xterm/addon-web-links @xterm/addon-webgl`
  - [ ] Import xterm.css
- [ ] Create useWebSocket hook
  - [ ] WebSocket connection management
  - [ ] Auto-reconnect on disconnect
  - [ ] Connection status (connecting/connected/disconnected)
  - [ ] Token passed via query parameter
- [ ] Create useTerminal hook
  - [ ] Terminal instance management
  - [ ] xterm.js setup with addons
  - [ ] Terminal resize handling
  - [ ] Message protocol handling (input/output/resize)
- [ ] Create Terminal component
  - [ ] Render terminal container
  - [ ] Initialize xterm.js instance
  - [ ] Connect to WebSocket
  - [ ] Handle input/output
  - [ ] Handle resize events
  - [ ] Clean up on unmount
- [ ] Create TerminalToolbar component
  - [ ] Reconnect button
  - [ ] Fullscreen button
  - [ ] Connection status indicator
  - [ ] Clear terminal button (optional)
- [ ] Integrate into WorkspaceDetailPage
  - [ ] Terminal tab
  - [ ] Load terminal on tab switch
  - [ ] Dispose terminal when leaving tab

#### 🔌 External Interfaces

**useWebSocket Hook**:
```typescript
// src/hooks/useWebSocket.ts
import { useEffect, useRef, useState } from 'react'
import { useAtom } from 'jotai'
import { tokenAtom } from '@/stores/auth'

export const useWebSocket = (url: string) => {
  const [token] = useAtom(tokenAtom)
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
        // Reconnect after 3 seconds
        setTimeout(connect, 3000)
      }

      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
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

**Terminal Component**:
```typescript
// src/components/terminal/Terminal.tsx
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
  const wsUrl = `ws://${window.location.host}/ws/terminal/${workspaceId}`
  const { ws, status } = useWebSocket(wsUrl)

  useEffect(() => {
    if (!terminalRef.current) return

    // Create terminal instance
    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
      },
    })

    // Load addons
    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()
    const webglAddon = new WebglAddon()

    term.loadAddon(fitAddon)
    term.loadAddon(webLinksAddon)
    try {
      term.loadAddon(webglAddon)
    } catch {
      // WebGL not supported, fallback to canvas
    }

    term.open(terminalRef.current)
    fitAddon.fit()

    termRef.current = term

    // WebSocket message handling
    if (ws && status === 'connected') {
      // Send user input to server
      term.onData((data) => {
        ws.send(JSON.stringify({ type: 'input', data }))
      })

      // Receive output from server
      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data)
          if (msg.type === 'output') {
            term.write(msg.data)
          } else if (msg.type === 'error') {
            term.write(`\r\n\x1b[31mError: ${msg.data}\x1b[0m\r\n`)
          }
        } catch {
          // Ignore parse errors
        }
      }

      // Handle window resize
      const handleResize = () => {
        fitAddon.fit()
        ws.send(JSON.stringify({
          type: 'resize',
          cols: term.cols,
          rows: term.rows,
        }))
      }

      window.addEventListener('resize', handleResize)
      return () => {
        window.removeEventListener('resize', handleResize)
        term.dispose()
      }
    }

    return () => {
      term.dispose()
    }
  }, [workspaceId, ws, status])

  return (
    <div className="h-full w-full bg-[#1e1e1e] rounded-lg overflow-hidden">
      {status !== 'connected' && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/50 text-white">
          {status === 'connecting' ? 'Connecting...' : 'Disconnected'}
        </div>
      )}
      <div ref={terminalRef} className="h-full" />
    </div>
  )
}
```

#### ✅ Acceptance Criteria

- [ ] Terminal renders correctly
- [ ] WebSocket connects successfully
- [ ] Can type commands and see output
- [ ] Terminal resize works
- [ ] Connection status displays correctly
- [ ] Auto-reconnect works after disconnect
- [ ] Fullscreen mode works
- [ ] Terminal disposes correctly on unmount
- [ ] Works with interactive programs (vim, top)

#### 📚 Dependencies

- Module 1, Module 2, Module 3
- `@xterm/xterm`, `@xterm/addon-fit`, `@xterm/addon-web-links`, `@xterm/addon-webgl`

---

### Module 7: Integration & Polish (集成与完善)

**Owner**: Agent 6, Agent 7
**Estimated Time**: 2-3 days
**Priority**: 🟡 Medium
**Dependencies**: All previous modules

#### 📦 Components

1. **Settings Page** (`src/pages/SettingsPage.tsx`)
2. **Error Boundary** (`src/components/ErrorBoundary.tsx`)
3. **Loading States** (global and component-level)
4. **Toast Notifications** (using sonner or react-hot-toast)

#### 📋 Task Checklist

**Agent 6 - Settings & Error Handling**:
- [ ] Create Settings Page
  - [ ] Account section (logout button)
  - [ ] Theme settings (terminal theme)
  - [ ] About section (version info)
- [ ] Implement Error Boundary
  - [ ] Catch React errors
  - [ ] Display friendly error message
  - [ ] Log errors to console
- [ ] Add global error handling
  - [ ] API error handling
  - [ ] Network error handling
  - [ ] Toast notifications for errors

**Agent 7 - Loading States & Polish**:
- [ ] Add loading states
  - [ ] Page-level loading spinner
  - [ ] Button loading states
  - [ ] Skeleton loaders for workspace cards
- [ ] Install and configure toast notifications
  - [ ] `npm install sonner` or `react-hot-toast`
  - [ ] Add toast container to App
  - [ ] Use toasts for success/error messages
- [ ] Polish UI/UX
  - [ ] Add transitions and animations
  - [ ] Improve responsive design
  - [ ] Add keyboard shortcuts (optional)
  - [ ] Improve accessibility (aria labels, focus management)

#### 🔌 External Interfaces

**Settings Page**:
```typescript
// src/pages/SettingsPage.tsx
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/hooks/useAuth'
import { authApi } from '@/api/auth'
import { useNavigate } from 'react-router-dom'

export const SettingsPage = () => {
  const { logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = async () => {
    try {
      await authApi.logout()
      logout()
      navigate('/login')
    } catch (error) {
      console.error('Logout failed:', error)
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Settings</h1>

      <Card>
        <CardHeader>
          <CardTitle>Account</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground mb-4">You are logged in</p>
          <Button onClick={handleLogout}>Logout</Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>About</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm">ViBox v1.2.0</p>
          <p className="text-sm text-muted-foreground">Backend: Go 1.25</p>
          <p className="text-sm text-muted-foreground">Frontend: React 18 + Vite</p>
        </CardContent>
      </Card>
    </div>
  )
}
```

**Error Boundary**:
```typescript
// src/components/ErrorBoundary.tsx
import { Component, ErrorInfo, ReactNode } from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('React Error:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-background">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>Something went wrong</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm text-muted-foreground">
                {this.state.error?.message || 'An unexpected error occurred'}
              </p>
              <Button onClick={() => window.location.reload()}>
                Reload Page
              </Button>
            </CardContent>
          </Card>
        </div>
      )
    }

    return this.props.children
  }
}
```

#### ✅ Acceptance Criteria

- [ ] Settings page works correctly
- [ ] Logout functionality works
- [ ] Error boundary catches errors
- [ ] Toast notifications show for success/error
- [ ] Loading states display correctly
- [ ] Responsive design works on all devices
- [ ] No console errors or warnings
- [ ] Accessibility features work (keyboard navigation, screen readers)

#### 📚 Dependencies

- All previous modules
- `sonner` or `react-hot-toast` for toast notifications

---

### Module 8: Testing & Optimization (测试与优化)

**Owner**: Agent 8
**Estimated Time**: 2-3 days
**Priority**: 🟢 Low (final stage)
**Dependencies**: All other modules

#### 📦 Components

1. **Unit Tests** (optional, using Vitest)
2. **Integration Tests** (manual or automated)
3. **Performance Optimization**
4. **Build Configuration**

#### 📋 Task Checklist

- [ ] Manual testing
  - [ ] Test all user flows (login → create workspace → terminal → delete)
  - [ ] Test error scenarios (invalid token, network errors, etc.)
  - [ ] Test on different browsers (Chrome, Firefox, Safari, Edge)
  - [ ] Test responsive design on different screen sizes
  - [ ] Test WebSocket reconnection
  - [ ] Test terminal with various commands
- [ ] Performance optimization
  - [ ] Code splitting (lazy load pages)
  - [ ] Optimize bundle size
  - [ ] Optimize images (if any)
  - [ ] Remove unused dependencies
  - [ ] Enable production build optimizations
- [ ] Build configuration
  - [ ] Configure build output directory
  - [ ] Configure asset chunking
  - [ ] Test production build locally
  - [ ] Verify source maps are disabled in production
- [ ] Documentation
  - [ ] Update README with frontend setup instructions
  - [ ] Document environment variables
  - [ ] Add troubleshooting guide (optional)

#### 🔌 External Interfaces

**Vite Build Config**:
```typescript
// vite.config.ts
export default defineConfig({
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'xterm-vendor': ['@xterm/xterm', '@xterm/addon-fit', '@xterm/addon-web-links'],
          'ui-vendor': ['jotai', 'axios'],
        },
      },
    },
  },
})
```

**Lazy Loading Pages**:
```typescript
// src/App.tsx
import { lazy, Suspense } from 'react'

const WorkspacesPage = lazy(() => import('./pages/WorkspacesPage'))
const WorkspaceDetailPage = lazy(() => import('./pages/WorkspaceDetailPage'))
const SettingsPage = lazy(() => import('./pages/SettingsPage'))

// Wrap routes with Suspense
const routes = [
  {
    path: '/',
    element: (
      <Suspense fallback={<LoadingSpinner />}>
        <WorkspacesPage />
      </Suspense>
    ),
  },
  // ... other routes
]
```

#### ✅ Acceptance Criteria

- [ ] All user flows work correctly
- [ ] No console errors or warnings
- [ ] Works on all major browsers
- [ ] Responsive design works on all screen sizes
- [ ] Production build succeeds
- [ ] Bundle size is optimized (<500KB gzipped)
- [ ] Page load time is fast (<2 seconds)
- [ ] WebSocket reconnection works reliably

#### 📚 Dependencies

- All other modules

---

## Development Schedule

### 📅 Detailed Timeline

#### Week 1 (Day 1-7)

**Day 1-2: Round 1**
- Agent 1: Complete Module 1 (Foundation Layer)
- Milestone: Project setup, Tailwind CSS, Router, Base Layout

**Day 3-4: Round 2**
- Agent 2: Complete Module 2 (State Management) + Module 3 (API Integration)
- Milestone: Jotai atoms, API client, Type definitions

**Day 5-7: Round 3 Start**
- Agent 3: Start Module 4 (Auth UI)
- Agent 4: Start Module 5 (Workspace UI)
- Agent 5: Start Module 6 (Terminal Integration)

#### Week 2 (Day 8-14)

**Day 8-11: Round 3 Continue**
- Agent 3, 4, 5: Continue parallel development of UI modules
- Milestone: Auth, Workspace UI, Terminal all complete

**Day 12-13: Round 4**
- Agent 6: Complete Settings + Error Handling
- Agent 7: Complete Loading States + Polish
- Milestone: Full integration, polished UI

**Day 14: Round 5**
- Agent 8: Testing & Optimization
- Milestone: Production-ready frontend

---

## Team Collaboration Guidelines

### 📢 Communication Mechanisms

1. **Interface First**: Each module defines clear interfaces upfront
2. **Mock Data**: Use mock data when dependencies are not ready
3. **Daily Sync**: Daily progress updates and blocker discussions
4. **Integration Testing**: Run integration tests at the end of each Round

### 🔧 Tooling

- **Code Repository**: Git + GitHub
- **Branching Strategy**:
  - `main` - Main branch
  - `module-1-foundation` - Module 1 development
  - `module-2-state` - Module 2 development
  - ... (one branch per module)
- **Pull Requests**: Submit PR after each module completion
- **Code Review**: At least one reviewer before merging

### ✅ Quality Assurance

- **Code Standards**: Use ESLint and Prettier
- **Type Safety**: All components and functions must be typed
- **Testing**: Manual testing for each module
- **Documentation**: Each module provides README and examples

---

## Risk Management

### ⚠️ Potential Risks

1. **Module Dependency Blocking**
   - Mitigation: Follow Round order strictly, use mocks for incomplete dependencies

2. **Interface Mismatch**
   - Mitigation: Review all interface definitions together after Round 1

3. **Integration Issues**
   - Mitigation: Run integration tests at the end of each Round

4. **Backend API Changes**
   - Mitigation: Backend v1.1.0 is stable, API spec is finalized

### 🎯 Success Factors

1. **Clear Interface Definitions**: Avoid rework later
2. **Strict Dependency Management**: Develop in order
3. **Sufficient Testing**: Manual testing + integration testing
4. **Continuous Integration**: Integrate modules immediately after completion

---

## Summary

By breaking frontend development into 8 independent modules, we can achieve:

- ✅ **Parallel Development**: Up to 3 agents working simultaneously (Round 3)
- ✅ **Risk Reduction**: Modules are independent with clear dependencies
- ✅ **Improved Quality**: Each module is independently tested
- ✅ **Faster Delivery**: Estimated 9-14 days (consistent with original plan)

**Next Steps**:
1. Form development team (assign Agents)
2. Review interface definitions
3. Start Round 1 development

---

**Document Version**: v1.0
**Created**: 2025-11-10
**Maintainer**: ViBox Team
