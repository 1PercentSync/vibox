import { useNavigate } from 'react-router-dom'
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import type { Workspace } from '@/api/types'

interface WorkspaceCardProps {
  workspace: Workspace
  onDelete: (id: string) => void
  onReset: (id: string) => void
}

export function WorkspaceCard({ workspace, onDelete, onReset }: WorkspaceCardProps) {
  const navigate = useNavigate()

  // Status configuration
  const statusConfig = {
    creating: {
      color: 'warning' as const,
      label: 'Creating...',
      canUseTerminal: false
    },
    running: {
      color: 'success' as const,
      label: 'Running',
      canUseTerminal: true
    },
    error: {
      color: 'destructive' as const,
      label: 'Error',
      canUseTerminal: !!workspace.container_id
    },
    failed: {
      color: 'destructive' as const,
      label: 'Failed',
      canUseTerminal: false
    },
  }

  const config = statusConfig[workspace.status]

  // Button availability
  const canUseTerminal = config.canUseTerminal
  const canUsePorts = workspace.status === 'running'

  const handleTerminal = () => {
    navigate(`/workspace/${workspace.id}`)
  }

  const handlePorts = () => {
    navigate(`/workspace/${workspace.id}?tab=ports`)
  }

  const handleOpenPort = (port: string) => {
    window.open(`/forward/${workspace.id}/${port}/`, '_blank')
  }

  return (
    <Card className="hover:shadow-lg transition-shadow duration-200">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle>{workspace.name}</CardTitle>
          <Badge variant={config.color}>{config.label}</Badge>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        <p className="text-sm text-muted-foreground">
          {workspace.config.image}
        </p>

        {/* Error message */}
        {workspace.error && (
          <Alert variant="destructive">
            <AlertDescription>{workspace.error}</AlertDescription>
          </Alert>
        )}

        {/* Port quick access (only for running workspaces with ports) */}
        {canUsePorts && workspace.ports && Object.keys(workspace.ports).length > 0 && (
          <div className="space-y-2">
            <p className="text-sm font-medium">Quick Access:</p>
            <div className="flex flex-wrap gap-2">
              {Object.entries(workspace.ports).map(([port, label]) => (
                <Button
                  key={port}
                  size="sm"
                  variant="outline"
                  onClick={() => handleOpenPort(port)}
                >
                  {label}:{port}
                </Button>
              ))}
            </div>
          </div>
        )}
      </CardContent>

      <CardFooter className="flex flex-wrap gap-2 pt-4 border-t border-border">
        <Button
          size="sm"
          disabled={!canUseTerminal}
          onClick={handleTerminal}
        >
          Terminal
        </Button>
        <Button
          size="sm"
          variant="outline"
          disabled={!canUsePorts}
          onClick={handlePorts}
        >
          Ports
        </Button>
        <Button
          size="sm"
          variant="outline"
          onClick={() => onReset(workspace.id)}
        >
          Reset
        </Button>
        <Button
          size="sm"
          variant="destructive"
          onClick={() => onDelete(workspace.id)}
        >
          Delete
        </Button>
      </CardFooter>
    </Card>
  )
}
