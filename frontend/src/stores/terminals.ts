import { atom } from 'jotai'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'

// Terminal instance cache to persist across page navigations
interface TerminalInstance {
  terminal: XTerm
  fitAddon: FitAddon
  workspaceId: string
  websocket: WebSocket | null
  createdAt: number
}

// Store terminal instances by workspace ID
export const terminalInstancesAtom = atom<Map<string, TerminalInstance>>(new Map())

// Get terminal instance for a workspace
export function getTerminalInstance(
  instances: Map<string, TerminalInstance>,
  workspaceId: string
): TerminalInstance | null {
  return instances.get(workspaceId) || null
}

// Store terminal instance
export function setTerminalInstance(
  instances: Map<string, TerminalInstance>,
  workspaceId: string,
  terminal: XTerm,
  fitAddon: FitAddon,
  websocket: WebSocket | null = null
): Map<string, TerminalInstance> {
  const newInstances = new Map(instances)
  const existing = newInstances.get(workspaceId)

  newInstances.set(workspaceId, {
    terminal,
    fitAddon,
    workspaceId,
    websocket: websocket || existing?.websocket || null,
    createdAt: existing?.createdAt || Date.now(),
  })

  return newInstances
}

// Update WebSocket for a terminal
export function updateTerminalWebSocket(
  instances: Map<string, TerminalInstance>,
  workspaceId: string,
  websocket: WebSocket
): Map<string, TerminalInstance> {
  const newInstances = new Map(instances)
  const existing = newInstances.get(workspaceId)

  if (existing) {
    newInstances.set(workspaceId, {
      ...existing,
      websocket,
    })
  }

  return newInstances
}

// Remove terminal instance (when workspace is deleted)
export function removeTerminalInstance(
  instances: Map<string, TerminalInstance>,
  workspaceId: string
): Map<string, TerminalInstance> {
  const newInstances = new Map(instances)
  const instance = newInstances.get(workspaceId)

  if (instance) {
    // Cleanup
    if (instance.websocket) {
      instance.websocket.close()
    }
    instance.terminal.dispose()
    newInstances.delete(workspaceId)
  }

  return newInstances
}

// Clean up old terminal instances (e.g., older than 1 hour)
export function cleanupOldTerminals(
  instances: Map<string, TerminalInstance>,
  maxAgeMs: number = 60 * 60 * 1000
): Map<string, TerminalInstance> {
  const newInstances = new Map(instances)
  const now = Date.now()

  newInstances.forEach((instance, workspaceId) => {
    if (now - instance.createdAt > maxAgeMs) {
      if (instance.websocket) {
        instance.websocket.close()
      }
      instance.terminal.dispose()
      newInstances.delete(workspaceId)
    }
  })

  return newInstances
}

