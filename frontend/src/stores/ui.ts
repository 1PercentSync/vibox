// UI State Management using Jotai
import { atom } from 'jotai'
import { atomWithStorage } from 'jotai/utils'

// Terminal theme (stored in localStorage)
export type TerminalTheme = 'dark' | 'light'

export const terminalThemeAtom = atomWithStorage<TerminalTheme>(
  'terminal-theme',
  'dark'
)

// Sidebar open/close state
export const sidebarOpenAtom = atom<boolean>(true)

// Create workspace dialog open state
export const createWorkspaceDialogOpenAtom = atom<boolean>(false)

// Delete confirmation dialog state
export interface DeleteConfirmState {
  open: boolean
  workspaceId: string | null
  workspaceName: string | null
}

export const deleteConfirmDialogAtom = atom<DeleteConfirmState>({
  open: false,
  workspaceId: null,
  workspaceName: null,
})

// Loading states
export const isLoadingAtom = atom<boolean>(false)
export const isCreatingWorkspaceAtom = atom<boolean>(false)
export const isDeletingWorkspaceAtom = atom<boolean>(false)

// Toast/notification state
export interface ToastState {
  open: boolean
  message: string
  type: 'success' | 'error' | 'info'
}

export const toastAtom = atom<ToastState>({
  open: false,
  message: '',
  type: 'info',
})
