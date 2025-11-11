import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import type { Workspace } from '@/api/types'
import { workspaceApi } from '@/api/workspaces'

interface ConfigTabProps {
  workspace: Workspace
  onUpdate: () => void
}

export function ConfigTab({ workspace, onUpdate }: ConfigTabProps) {
  const [resetting, setResetting] = useState(false)

  const handleReset = async () => {
    if (
      !confirm(
        'Are you sure you want to reset this workspace? This will delete the current container and recreate it with the original configuration. All data in the container will be lost.'
      )
    ) {
      return
    }

    try {
      setResetting(true)
      await workspaceApi.reset(workspace.id)
      onUpdate()
    } catch (error) {
      console.error('Failed to reset workspace:', error)
      alert('Failed to reset workspace')
    } finally {
      setResetting(false)
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString()
  }

  return (
    <div className="space-y-6">
      {/* Basic Information */}
      <Card>
        <CardHeader>
          <CardTitle>Workspace Information</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div className="font-medium">Workspace ID:</div>
            <div className="text-muted-foreground font-mono">{workspace.id}</div>

            <div className="font-medium">Name:</div>
            <div className="text-muted-foreground">{workspace.name}</div>

            <div className="font-medium">Status:</div>
            <div className="text-muted-foreground">
              <span className="capitalize">{workspace.status}</span>
            </div>

            {workspace.container_id && (
              <>
                <div className="font-medium">Container ID:</div>
                <div className="text-muted-foreground font-mono">
                  {workspace.container_id}
                </div>
              </>
            )}

            <div className="font-medium">Image:</div>
            <div className="text-muted-foreground">{workspace.config.image}</div>

            <div className="font-medium">Created:</div>
            <div className="text-muted-foreground">
              {formatDate(workspace.created_at)}
            </div>
          </div>

          {workspace.error && (
            <Alert variant="destructive">
              <AlertDescription>{workspace.error}</AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>

      {/* Initialization Scripts */}
      {workspace.config.scripts.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Initialization Scripts</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {workspace.config.scripts.map((script) => (
              <div key={script.name} className="space-y-2">
                <div className="flex items-center justify-between">
                  <h4 className="font-medium">
                    {script.order}. {script.name}
                  </h4>
                </div>
                <pre className="p-3 bg-muted rounded-lg text-sm overflow-x-auto">
                  <code>{script.content}</code>
                </pre>
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      {/* Port Labels */}
      {workspace.ports && Object.keys(workspace.ports).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Port Labels</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {Object.entries(workspace.ports).map(([port, label]) => (
                <div key={port} className="flex items-center justify-between text-sm">
                  <span className="font-medium">{label}</span>
                  <span className="text-muted-foreground">Port {port}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Reset Workspace Section */}
      <Card>
        <CardHeader>
          <CardTitle>Reset Workspace</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <Alert>
            <AlertDescription>
              Resetting the workspace will delete the current container and
              recreate it with the original configuration. All data in the
              container will be lost. The initialization scripts will be
              re-executed.
            </AlertDescription>
          </Alert>
          <Button
            variant="destructive"
            onClick={handleReset}
            disabled={resetting}
          >
            {resetting ? 'Resetting...' : 'Reset Workspace'}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
