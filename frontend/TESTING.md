# Frontend Testing Guide

## Quick Start (Without Real Backend)

### 1. Start Mock Backend Server

In terminal 1:
```bash
cd frontend
npm run mock
```

This will start a mock backend server on **http://localhost:3000**

### 2. Start Frontend Dev Server

In terminal 2:
```bash
cd frontend
npm run dev
```

This will start the frontend on **http://localhost:5173**

### 3. Login

Open http://localhost:5173 in your browser.

**Use one of these test tokens:**
- `test`
- `demo`

Just type `test` in the token field and click Login!

---

## Mock Backend Features

The mock server provides all the necessary API endpoints:

### Authentication
- ✅ `POST /api/auth/login` - Login (accepts any token: "test", "demo", or any string)
- ✅ `POST /api/auth/logout` - Logout

### Workspaces
- ✅ `GET /api/workspaces` - List all workspaces
- ✅ `GET /api/workspaces/:id` - Get workspace details
- ✅ `POST /api/workspaces` - Create new workspace
- ✅ `DELETE /api/workspaces/:id` - Delete workspace
- ✅ `PUT /api/workspaces/:id/ports` - Update port mappings
- ✅ `POST /api/workspaces/:id/reset` - Reset workspace

---

## Testing Checklist

### ✅ Login Page
- [ ] Page loads with styled components
- [ ] Can enter token
- [ ] Enter key works
- [ ] Login with "test" token succeeds
- [ ] Shows success toast notification
- [ ] Redirects to workspaces page

### ✅ Workspaces Page
- [ ] Shows empty state when no workspaces
- [ ] Can create new workspace
- [ ] Shows workspace cards
- [ ] Workspace status badges display correctly
- [ ] Can delete workspace
- [ ] Can reset workspace
- [ ] Auto-refresh works (every 5 seconds)

### ✅ Settings Page
- [ ] Navigate to Settings from header
- [ ] Shows account information
- [ ] Logout button works
- [ ] Shows About section with version info

### ✅ UI/UX Features
- [ ] Toast notifications appear for all actions
- [ ] Skeleton loaders show during initial load
- [ ] All buttons have hover effects
- [ ] Responsive design works on mobile
- [ ] Error messages display correctly

---

## Known Limitations

Since this is a mock backend:

1. **No Terminal WebSocket** - Terminal functionality won't work
2. **No Port Forwarding** - `/forward/*` routes won't work
3. **In-Memory Storage** - Workspaces are lost when mock server restarts
4. **No Container Creation** - Workspaces are instantly "created"

To test these features, you need to run the real Go backend.

---

## Troubleshooting

### Styles Not Loading

If the page looks unstyled:

1. Clear browser cache (Ctrl+Shift+R or Cmd+Shift+R)
2. Check browser console for errors
3. Restart dev server: `Ctrl+C` then `npm run dev`

### Mock Server Port Already in Use

If you see "port 3000 already in use":

```bash
# On Windows
netstat -ano | findstr :3000
taskkill /PID <PID> /F

# On Mac/Linux
lsof -ti:3000 | xargs kill
```

Or change the port in `mock-server.js`:
```javascript
const PORT = 3001  // Change to any available port
```

### API Calls Failing

Check that:
1. Mock server is running on http://localhost:3000
2. Frontend proxy is configured in `vite.config.ts`
3. Browser network tab shows requests going to `/api/*`

---

## Next Steps

Once frontend testing is complete:

1. **Backend Integration**: Run the real Go backend
2. **Terminal Testing**: Test WebSocket terminal connections
3. **Port Forwarding**: Test container port access
4. **End-to-End Testing**: Full workflow testing

---

## Production Build

To test production build:

```bash
npm run build
npm run preview
```

This will:
1. Build optimized bundle
2. Start preview server on http://localhost:4173

---

**Happy Testing! 🎉**
