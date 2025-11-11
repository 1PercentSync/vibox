# ViBox Frontend

> React + TypeScript + Vite frontend for ViBox workspace management system

## Overview

The ViBox frontend is a modern, responsive web application built with React 18 and TypeScript. It provides a user-friendly interface for managing Docker-based workspaces, including real-time terminal access, port management, and workspace configuration.

## Tech Stack

- **Framework**: React 18.3+ (Function Components + Hooks)
- **Build Tool**: Vite 7.2+ (Fast HMR & Optimized Builds)
- **Language**: TypeScript 5.9+ (Full Type Safety)
- **Styling**: Tailwind CSS 4.1+ (Utility-First CSS)
- **UI Components**: shadcn UI (Radix UI Primitives)
- **State Management**: Jotai 2.15+ (Atomic State)
- **Routing**: React Router DOM 7.9+
- **HTTP Client**: Axios 1.13+
- **Terminal**: xterm.js 5.5+ (WebGL Rendering)
- **Notifications**: Sonner 2.0+ (Toast Notifications)
- **Icons**: Lucide React

## Features

- ✅ **User Authentication** - Token-based login with Cookie session management
- ✅ **Workspace Management** - Create, delete, reset workspaces with real-time status
- ✅ **Web Terminal** - Full-featured terminal with xterm.js and WebSocket
- ✅ **Port Management** - Quick access buttons and port forwarding UI
- ✅ **Responsive Design** - Works on desktop, tablet, and mobile
- ✅ **Error Handling** - Global error boundary and toast notifications
- ✅ **Loading States** - Skeleton loaders for better UX
- ✅ **Code Splitting** - Lazy-loaded pages for optimal performance

## Project Structure

```
frontend/
├── src/
│   ├── api/                    # API client and type definitions
│   │   ├── client.ts          # Axios instance with interceptors
│   │   ├── auth.ts            # Authentication API
│   │   ├── workspaces.ts      # Workspace API
│   │   └── types.ts           # API type definitions
│   ├── components/            # React components
│   │   ├── ui/                # shadcn UI components
│   │   ├── layout/            # Layout components (Header, Sidebar, Layout)
│   │   ├── workspace/         # Workspace-related components
│   │   ├── terminal/          # Terminal components
│   │   ├── ErrorBoundary.tsx  # Error boundary
│   │   ├── LoadingSpinner.tsx # Loading indicator
│   │   └── ProtectedRoute.tsx # Route guard
│   ├── hooks/                 # Custom React hooks
│   │   ├── useAuth.ts         # Authentication hook
│   │   ├── useWorkspaces.ts   # Workspace management hook
│   │   └── useWebSocket.ts    # WebSocket connection hook
│   ├── pages/                 # Page components
│   │   ├── LoginPage.tsx      # Login page
│   │   ├── WorkspacesPage.tsx # Workspace list page
│   │   ├── WorkspaceDetailPage.tsx # Workspace detail page
│   │   └── SettingsPage.tsx   # Settings page
│   ├── stores/                # Jotai atoms (state management)
│   │   ├── auth.ts            # Authentication state
│   │   ├── workspaces.ts      # Workspace state
│   │   └── ui.ts              # UI state
│   ├── types/                 # TypeScript type definitions
│   │   └── workspace.ts       # Workspace types
│   ├── lib/                   # Utility functions
│   │   └── utils.ts           # Class name utilities
│   ├── App.tsx                # Root component with router
│   ├── main.tsx               # Application entry point
│   └── index.css              # Global styles
├── public/                    # Static assets
├── mock-server.js             # Mock backend for development
├── package.json               # Dependencies and scripts
├── vite.config.ts             # Vite configuration
├── tailwind.config.ts         # Tailwind CSS configuration
├── tsconfig.json              # TypeScript configuration
└── components.json            # shadcn UI configuration
```

## Getting Started

### Prerequisites

- Node.js 18+
- npm or pnpm
- Backend running on `http://localhost:3000` (or use mock server)

### Installation

```bash
# Install dependencies
npm install
```

### Development

#### With Real Backend

```bash
# 1. Start the Go backend (in another terminal)
cd ..
export API_TOKEN=dev-token-123
go run ./cmd/server

# 2. Start frontend dev server
npm run dev

# Frontend runs on http://localhost:5173
# API requests are proxied to http://localhost:3000
```

#### With Mock Backend

```bash
# 1. Start mock server (in one terminal)
npm run mock

# 2. Start frontend dev server (in another terminal)
npm run dev
```

### Production Build

```bash
# Build for production
npm run build

# Output: dist/
# Bundle size: ~250KB gzipped (initial load)

# Preview production build
npm run preview
# Opens http://localhost:4173
```

## Scripts

| Script | Description |
|--------|-------------|
| `npm run dev` | Start Vite dev server (HMR enabled) |
| `npm run build` | Build for production (TypeScript + Vite) |
| `npm run preview` | Preview production build |
| `npm run lint` | Run ESLint |
| `npm run mock` | Start mock backend server |

## Environment Variables

The frontend uses Vite's proxy configuration, no environment variables needed for development.

For production deployment, ensure the backend API is accessible at the same origin (or configure CORS).

## API Integration

The frontend communicates with the backend via:

1. **RESTful API** (`/api/*`) - Workspace management, authentication
2. **WebSocket** (`/ws/*`) - Real-time terminal communication
3. **Port Forwarding** (`/forward/*`) - Access container HTTP services

All requests use **Cookie-based authentication** (HttpOnly Cookie set by backend).

See [API Specification](../docs/API_SPECIFICATION.md) for details.

## State Management

We use **Jotai** for state management:

- **Atomic state** - Each piece of state is an independent atom
- **Derived state** - Computed values with automatic dependency tracking
- **Minimal boilerplate** - No actions, reducers, or providers
- **TypeScript-first** - Full type inference

Example:
```typescript
// stores/auth.ts
export const tokenAtom = atom<string | null>(
  localStorage.getItem('api_token')
)

export const isAuthenticatedAtom = atom(
  (get) => get(tokenAtom) !== null
)

// In component
const [isAuthenticated] = useAtom(isAuthenticatedAtom)
```

## Performance Optimizations

1. **Code Splitting** - Lazy-loaded pages with React.lazy()
2. **Bundle Splitting** - Vendor chunks (React, xterm, UI libraries)
3. **Tree Shaking** - Unused code eliminated by Vite
4. **CSS Purging** - Tailwind CSS purges unused styles
5. **Minification** - Esbuild minifies JS and CSS
6. **Gzip Compression** - All assets served compressed

**Bundle Size**:
- React vendor: 31.57 KB gzipped
- xterm vendor: 98.56 KB gzipped
- UI vendor: 27.25 KB gzipped
- Total: ~250 KB gzipped (initial load)

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

WebSocket and xterm.js require modern browser features.

## Development Workflow

### Adding a New Page

1. Create page component in `src/pages/`
2. Add lazy import in `App.tsx`
3. Add route with `<Suspense>` wrapper
4. Add navigation link in Header/Sidebar

### Adding shadcn UI Component

```bash
# Example: add Dialog component
npx shadcn-ui@latest add dialog
```

### API Integration

1. Define types in `src/api/types.ts`
2. Add API methods in `src/api/workspaces.ts` (or new file)
3. Use in components with Axios (errors handled globally)

### State Management

1. Define atom in `src/stores/`
2. Use with `useAtom()` hook in components
3. For complex state, create custom hook in `src/hooks/`

## Testing

Currently no automated tests. Testing is done manually:

- ✅ User flows (login, create workspace, terminal, delete)
- ✅ Error scenarios (invalid token, network errors)
- ✅ Browser compatibility (Chrome, Firefox, Safari)
- ✅ Responsive design (desktop, tablet, mobile)

Future: Add Vitest for unit tests and Playwright for E2E tests.

## Troubleshooting

### Dev server not starting

```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

### API requests failing (404)

- Ensure backend is running on `http://localhost:3000`
- Check Vite proxy config in `vite.config.ts`
- Use mock server: `npm run mock`

### WebSocket connection fails

- Check backend WebSocket endpoint: `ws://localhost:3000/ws/terminal/:id`
- Ensure token is passed correctly
- Check browser console for errors

### Build errors

```bash
# Check TypeScript errors
npx tsc --noEmit

# Clear Vite cache
rm -rf node_modules/.vite
```

## Module Development

The frontend was developed in 8 modules:

1. ✅ **Module 1**: Foundation Layer (Vite + React + Tailwind + shadcn UI)
2. ✅ **Module 2**: State Management (Jotai atoms and custom hooks)
3. ✅ **Module 3**: API Integration (Axios client, auth, workspace APIs)
4. ✅ **Module 4**: Authentication UI (Login page, protected routes)
5. ✅ **Module 5**: Workspace UI (List, create, delete, port management)
6. ✅ **Module 6**: Terminal Integration (xterm.js + WebSocket)
7. ✅ **Module 7**: Integration & Polish (Settings, error handling, toasts)
8. ✅ **Module 8**: Testing & Optimization (Code splitting, build config)

See [Phase 2 Task Breakdown](../docs/PHASE2_TASK_BREAKDOWN.md) for detailed module descriptions.

## Contributing

When contributing to frontend:

1. Follow existing code style (TypeScript, functional components)
2. Use shadcn UI components when possible
3. Add types for all API responses
4. Test on multiple browsers
5. Keep bundle size small (check with `npm run build`)

## Documentation

- [Phase 2 Frontend Development](../docs/PHASE2_FRONTEND.md) - Technical overview
- [Task Breakdown](../docs/PHASE2_TASK_BREAKDOWN.md) - Module descriptions
- [Module Completion Reports](../docs/) - Implementation reports
- [API Specification](../docs/API_SPECIFICATION.md) - Backend API docs

## License

TBD

## Credits

Built with ❤️ using:
- [React](https://react.dev/)
- [Vite](https://vitejs.dev/)
- [Tailwind CSS](https://tailwindcss.com/)
- [shadcn/ui](https://ui.shadcn.com/)
- [xterm.js](https://xtermjs.org/)
- [Jotai](https://jotai.org/)
