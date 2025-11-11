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
