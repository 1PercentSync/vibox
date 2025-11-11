# ViBox Frontend

React + TypeScript + Vite frontend for ViBox workspace management platform.

## Tech Stack

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite 7
- **Styling**: Tailwind CSS
- **UI Components**: shadcn/ui
- **Routing**: React Router v6
- **State Management**: Jotai
- **HTTP Client**: Axios

## Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── ui/              # shadcn UI components
│   │   ├── layout/          # Layout components (Header, Sidebar)
│   │   └── ProtectedRoute.tsx
│   ├── pages/               # Page components
│   ├── stores/              # Jotai state stores
│   ├── lib/                 # Utilities
│   ├── hooks/               # Custom hooks
│   ├── App.tsx              # Root component with routes
│   └── main.tsx             # Entry point
├── public/
├── index.html
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
├── tsconfig.json
└── package.json
```

## Development

### Prerequisites

- Node.js 18+
- npm or yarn

### Install Dependencies

```bash
npm install
```

### Start Development Server

```bash
npm run dev
```

The app will be available at `http://localhost:5173/`

### Build for Production

```bash
npm run build
```

### Preview Production Build

```bash
npm run preview
```

## Features

- ✅ Authentication with route protection
- ✅ Workspace management UI
- ✅ Responsive design with Tailwind CSS
- ✅ Modern component library (shadcn/ui)
- ✅ API proxy configuration for backend
- ✅ TypeScript for type safety

## API Proxy

The Vite dev server is configured to proxy API requests to the backend:

- `/api/*` → `http://localhost:3000/api/*`
- `/ws/*` → `ws://localhost:3000/ws/*`
- `/forward/*` → `http://localhost:3000/forward/*`

## Environment Variables

Create a `.env` file based on `.env.example`:

```env
VITE_API_URL=http://localhost:3000
```

## Module 1 Completion Status

### ✅ Completed Tasks

1. ✅ Project initialized with Vite + React + TypeScript
2. ✅ Tailwind CSS installed and configured
3. ✅ shadcn/ui components installed (Button, Card, Input, Dialog, Badge)
4. ✅ React Router configured with protected routes
5. ✅ Base layout components created (Layout, Header, Sidebar)
6. ✅ Development environment configured (ESLint, .env, proxy)
7. ✅ All acceptance criteria verified

### Verification Results

- ✅ Project starts successfully (`npm run dev`)
- ✅ TypeScript compiles without errors
- ✅ Tailwind CSS styles applied
- ✅ shadcn UI components working
- ✅ Routes configured correctly
- ✅ API proxy configured
- ✅ Hot module reload working

## Next Steps

Continue with Module 2: Authentication Module development.
