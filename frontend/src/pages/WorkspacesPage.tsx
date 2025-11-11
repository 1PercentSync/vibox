import { useState } from 'react'
import { useAtomValue } from 'jotai'
import { Button } from '@/components/ui/button'
import { WorkspaceCard } from '@/components/workspace/WorkspaceCard'
import { CreateWorkspaceDialog } from '@/components/workspace/CreateWorkspaceDialog'
import { DeleteConfirmDialog } from '@/components/workspace/DeleteConfirmDialog'
import { useWorkspaces } from '@/hooks/useWorkspaces'
import { isLoadingAtom } from '@/stores/ui'
import { workspaceApi } from '@/api/workspaces'
import type { CreateWorkspaceRequest } from '@/api/types'

export function WorkspacesPage() {
  const { workspaces, refetch } = useWorkspaces({ autoRefresh: true, interval: 5000 })
  const isLoading = useAtomValue(isLoadingAtom)

  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [selectedWorkspaceId, setSelectedWorkspaceId] = useState<string | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  const selectedWorkspace = workspaces.find((w) => w.id === selectedWorkspaceId)

  const handleCreateWorkspace = async (data: CreateWorkspaceRequest) => {
    await workspaceApi.create(data)
    refetch()
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
      refetch()
      setDeleteDialogOpen(false)
      setSelectedWorkspaceId(null)
    } catch (error) {
      console.error('Failed to delete workspace:', error)
    } finally {
      setDeleteLoading(false)
    }
  }

  const handleResetWorkspace = async (id: string) => {
    if (!confirm('Are you sure you want to reset this workspace? All data will be lost.')) {
      return
    }

    try {
      await workspaceApi.reset(id)
      refetch()
    } catch (error) {
      console.error('Failed to reset workspace:', error)
      alert('Failed to reset workspace')
    }
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Workspaces</h1>
          <p className="text-muted-foreground">
            Manage your containerized development environments
          </p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          + New Workspace
        </Button>
      </div>

      {/* Loading State */}
      {isLoading && workspaces.length === 0 && (
        <div className="text-center py-12">
          <p className="text-muted-foreground">Loading workspaces...</p>
        </div>
      )}

      {/* Empty State */}
      {!isLoading && workspaces.length === 0 && (
        <div className="text-center py-12 space-y-3">
          <p className="text-lg font-medium">No workspaces yet</p>
          <p className="text-muted-foreground">
            Create your first workspace to get started
          </p>
          <Button onClick={() => setCreateDialogOpen(true)}>
            + Create Workspace
          </Button>
        </div>
      )}

      {/* Workspaces Grid */}
      {workspaces.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {workspaces.map((workspace) => (
            <WorkspaceCard
              key={workspace.id}
              workspace={workspace}
              onDelete={handleDeleteClick}
              onReset={handleResetWorkspace}
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
    </div>
  )
}
