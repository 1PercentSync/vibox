import { useState } from 'react'
import { useAtomValue } from 'jotai'
import { Button } from '@/components/ui/button'
import { WorkspaceCard } from '@/components/workspace/WorkspaceCard'
import { WorkspaceCardSkeleton } from '@/components/workspace/WorkspaceCardSkeleton'
import { CreateWorkspaceDialog } from '@/components/workspace/CreateWorkspaceDialog'
import { DeleteConfirmDialog } from '@/components/workspace/DeleteConfirmDialog'
import { ResetConfirmDialog } from '@/components/workspace/ResetConfirmDialog'
import { useWorkspaces } from '@/hooks/useWorkspaces'
import { isLoadingAtom } from '@/stores/ui'
import { workspaceApi } from '@/api/workspaces'
import { toast } from 'sonner'
import type { CreateWorkspaceRequest } from '@/api/types'

export function WorkspacesPage() {
  const { workspaces, refetch } = useWorkspaces({ autoRefresh: true, interval: 5000 })
  const isLoading = useAtomValue(isLoadingAtom)

  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [resetDialogOpen, setResetDialogOpen] = useState(false)
  const [selectedWorkspaceId, setSelectedWorkspaceId] = useState<string | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)
  const [resetLoading, setResetLoading] = useState(false)

  const selectedWorkspace = workspaces.find((w) => w.id === selectedWorkspaceId)

  const handleCreateWorkspace = async (data: CreateWorkspaceRequest) => {
    try {
      await workspaceApi.create(data)
      toast.success('Workspace created successfully')
      refetch()
    } catch (error) {
      // Error already handled by axios interceptor
      console.error('Failed to create workspace:', error)
    }
  }

  const handleDeleteClick = (id: string) => {
    setSelectedWorkspaceId(id)
    setDeleteDialogOpen(true)
  }

  const handleDeleteConfirm = async () => {
    if (!selectedWorkspaceId) return

    try {
      setDeleteLoading(true)
      await workspaceApi.delete(selectedWorkspaceId)
      toast.success('Workspace deleted successfully')
      refetch()
      setDeleteDialogOpen(false)
      setSelectedWorkspaceId(null)
    } catch (error) {
      // Error already handled by axios interceptor
      console.error('Failed to delete workspace:', error)
    } finally {
      setDeleteLoading(false)
    }
  }

  const handleResetClick = (id: string) => {
    setSelectedWorkspaceId(id)
    setResetDialogOpen(true)
  }

  const handleResetConfirm = async () => {
    if (!selectedWorkspaceId) return

    try {
      setResetLoading(true)
      await workspaceApi.reset(selectedWorkspaceId)
      toast.success('Workspace reset successfully')
      refetch()
      setResetDialogOpen(false)
      setSelectedWorkspaceId(null)
    } catch (error) {
      // Error already handled by axios interceptor
      console.error('Failed to reset workspace:', error)
    } finally {
      setResetLoading(false)
    }
  }

  return (
    <div className="container mx-auto px-6 py-8 space-y-8 max-w-7xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Workspaces</h1>
          <p className="text-muted-foreground mt-1">
            Manage your containerized development environments
          </p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          + New Workspace
        </Button>
      </div>

      {/* Loading State */}
      {isLoading && workspaces.length === 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {Array.from({ length: 3 }).map((_, index) => (
            <WorkspaceCardSkeleton key={index} />
          ))}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && workspaces.length === 0 && (
        <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-3">
          <p className="text-lg font-medium">No workspaces yet</p>
          <p className="text-muted-foreground">
            Create your first workspace to get started
          </p>
        </div>
      )}

      {/* Workspaces Grid */}
      {workspaces.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {workspaces.map((workspace) => (
            <WorkspaceCard
              key={workspace.id}
              workspace={workspace}
              onDelete={handleDeleteClick}
              onReset={handleResetClick}
            />
          ))}
        </div>
      )}

      {/* Create Dialog */}
      <CreateWorkspaceDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSubmit={handleCreateWorkspace}
      />

      {/* Delete Confirmation Dialog */}
      <DeleteConfirmDialog
        open={deleteDialogOpen}
        workspaceName={selectedWorkspace?.name}
        onOpenChange={setDeleteDialogOpen}
        onConfirm={handleDeleteConfirm}
        loading={deleteLoading}
      />

      {/* Reset Confirmation Dialog */}
      <ResetConfirmDialog
        open={resetDialogOpen}
        workspaceName={selectedWorkspace?.name}
        onOpenChange={setResetDialogOpen}
        onConfirm={handleResetConfirm}
        loading={resetLoading}
      />
    </div>
  )
}
