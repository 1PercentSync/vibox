import express from 'express'
import cors from 'cors'

const app = express()
app.use(cors())
app.use(express.json())

// Mock data
let workspaces = []
let authToken = 'test-token-123'

// Auth endpoints
app.post('/api/auth/login', (req, res) => {
  const { token } = req.body
  if (token === authToken || token === 'test' || token === 'demo') {
    res.cookie('vibox-token', token, { httpOnly: true, maxAge: 86400000 })
    res.json({ message: 'Login successful' })
  } else {
    res.status(401).json({ error: 'Invalid token', code: 'UNAUTHORIZED' })
  }
})

app.post('/api/auth/logout', (req, res) => {
  res.clearCookie('vibox-token')
  res.json({ message: 'Logout successful' })
})

// Workspace endpoints
app.get('/api/workspaces', (req, res) => {
  res.json(workspaces)
})

app.get('/api/workspaces/:id', (req, res) => {
  const workspace = workspaces.find(w => w.id === req.params.id)
  if (workspace) {
    res.json(workspace)
  } else {
    res.status(404).json({ error: 'Workspace not found', code: 'NOT_FOUND' })
  }
})

app.post('/api/workspaces', (req, res) => {
  const { name, image = 'ubuntu:22.04', scripts = [], ports = {} } = req.body
  const workspace = {
    id: `ws-${Date.now()}`,
    name,
    container_id: `container-${Date.now()}`,
    status: 'running',
    created_at: new Date().toISOString(),
    config: {
      image,
      scripts
    },
    ports
  }
  workspaces.push(workspace)
  res.status(201).json(workspace)
})

app.delete('/api/workspaces/:id', (req, res) => {
  const index = workspaces.findIndex(w => w.id === req.params.id)
  if (index !== -1) {
    workspaces.splice(index, 1)
    res.json({ message: 'Workspace deleted successfully', id: req.params.id })
  } else {
    res.status(404).json({ error: 'Workspace not found', code: 'NOT_FOUND' })
  }
})

app.put('/api/workspaces/:id/ports', (req, res) => {
  const workspace = workspaces.find(w => w.id === req.params.id)
  if (workspace) {
    workspace.ports = req.body.ports
    res.json(workspace)
  } else {
    res.status(404).json({ error: 'Workspace not found', code: 'NOT_FOUND' })
  }
})

app.post('/api/workspaces/:id/reset', (req, res) => {
  const workspace = workspaces.find(w => w.id === req.params.id)
  if (workspace) {
    workspace.status = 'creating'
    setTimeout(() => {
      workspace.status = 'running'
    }, 2000)
    res.json({
      message: 'Workspace reset successfully',
      workspace
    })
  } else {
    res.status(404).json({ error: 'Workspace not found', code: 'NOT_FOUND' })
  }
})

const PORT = 3000
app.listen(PORT, () => {
  console.log(`Mock backend server running on http://localhost:${PORT}`)
  console.log(`Use token: "test" or "demo" to login`)
})
