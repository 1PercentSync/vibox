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
