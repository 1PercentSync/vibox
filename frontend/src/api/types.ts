// API Type Definitions for ViBox Frontend
// Based on API_SPECIFICATION.md

export type WorkspaceStatus = 'creating' | 'running' | 'error' | 'failed'

export interface Script {
  name: string
  content: string
  order: number
}

export interface WorkspaceConfig {
  image: string
  scripts: Script[]
}

export interface Workspace {
  id: string
  name: string
  container_id?: string
  status: WorkspaceStatus
  created_at: string
  config: WorkspaceConfig
  ports?: Record<string, string>
  error?: string
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

export interface ResetWorkspaceResponse {
  message: string
  workspace: Workspace
}

export interface LoginRequest {
  token: string
}

export interface LoginResponse {
  message: string
}

export interface LogoutResponse {
  message: string
}

export interface DeleteWorkspaceResponse {
  message: string
  id: string
}
