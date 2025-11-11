// Workspace State Management using Jotai
import { atom } from 'jotai'
import type { Workspace } from '@/api/types'

// Atom to store all workspaces
export const workspacesAtom = atom<Workspace[]>([])

// Atom to store the currently selected workspace ID
export const selectedWorkspaceIdAtom = atom<string | null>(null)

// Derived atom to get the currently selected workspace
export const selectedWorkspaceAtom = atom((get) => {
  const workspaces = get(workspacesAtom)
  const selectedId = get(selectedWorkspaceIdAtom)
  return workspaces.find((ws) => ws.id === selectedId) || null
})

// Derived atom to check if any workspace is creating
export const hasCreatingWorkspacesAtom = atom((get) => {
  const workspaces = get(workspacesAtom)
  return workspaces.some((ws) => ws.status === 'creating')
})

// Derived atom to get running workspaces count
export const runningWorkspacesCountAtom = atom((get) => {
  const workspaces = get(workspacesAtom)
  return workspaces.filter((ws) => ws.status === 'running').length
})

// Derived atom to get error workspaces
export const errorWorkspacesAtom = atom((get) => {
  const workspaces = get(workspacesAtom)
  return workspaces.filter((ws) => ws.status === 'error' || ws.status === 'failed')
})
