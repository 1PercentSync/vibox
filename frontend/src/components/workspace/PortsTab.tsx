import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Alert, AlertDescription } from '@/components/ui/alert'
import type { Workspace } from '@/api/types'
import { workspaceApi } from '@/api/workspaces'

interface PortsTabProps {
  workspace: Workspace
  onUpdate: () => void
}

export function PortsTab({ workspace, onUpdate }: PortsTabProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editPorts, setEditPorts] = useState<Record<string, string>>(
    workspace.ports || {}
  )
  const [newPort, setNewPort] = useState('')
  const [newLabel, setNewLabel] = useState('')
  const [anyPort, setAnyPort] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const handleAddPort = () => {
    if (!newPort.trim() || !newLabel.trim()) {
      setError('Both port and label are required')
      return
    }

    const portNum = parseInt(newPort.trim())
    if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
      setError('Invalid port number')
      return
    }

    setEditPorts({
      ...editPorts,
      [newPort.trim()]: newLabel.trim(),
    })
    setNewPort('')
    setNewLabel('')
    setError('')
  }

  const handleRemovePort = (port: string) => {
    const updated = { ...editPorts }
    delete updated[port]
    setEditPorts(updated)
  }

  const handleSave = async () => {
    try {
      setSaving(true)
      setError('')
      await workspaceApi.updatePorts(workspace.id, { ports: editPorts })
      onUpdate()
      setIsEditing(false)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update ports')
    } finally {
      setSaving(false)
    }
  }

  const handleCancel = () => {
    setEditPorts(workspace.ports || {})
    setIsEditing(false)
    setError('')
  }

  const handleOpenAnyPort = () => {
    const portNum = parseInt(anyPort.trim())
    if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
      setError('Invalid port number')
      return
    }

    window.open(`/forward/${workspace.id}/${portNum}/`, '_blank')
    setAnyPort('')
    setError('')
  }

  const currentPorts = isEditing ? editPorts : workspace.ports || {}

  return (
    <div className="space-y-6">
      {/* Saved Ports Section */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">Saved Port Labels</h3>
          {!isEditing && (
            <Button onClick={() => setIsEditing(true)} size="sm">
              Edit Ports
            </Button>
          )}
        </div>

        {Object.keys(currentPorts).length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No saved ports. Click "Edit Ports" to add port labels.
          </p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Service Name</TableHead>
                <TableHead>Port</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {Object.entries(currentPorts).map(([port, label]) => (
                <TableRow key={port}>
                  <TableCell className="font-medium">{label}</TableCell>
                  <TableCell>{port}</TableCell>
                  <TableCell className="text-right space-x-2">
                    {!isEditing && workspace.status === 'running' && (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() =>
                          window.open(
                            `/forward/${workspace.id}/${port}/`,
                            '_blank'
                          )
                        }
                      >
                        Open
                      </Button>
                    )}
                    {isEditing && (
                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={() => handleRemovePort(port)}
                      >
                        Remove
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}

        {/* Edit Mode: Add New Port */}
        {isEditing && (
          <div className="space-y-3 p-4 border rounded-lg">
            <h4 className="font-medium">Add New Port Label</h4>
            <div className="flex gap-2">
              <div className="w-32">
                <Label htmlFor="new-port">Port</Label>
                <Input
                  id="new-port"
                  placeholder="8080"
                  value={newPort}
                  onChange={(e) => setNewPort(e.target.value)}
                />
              </div>
              <div className="flex-1">
                <Label htmlFor="new-label">Service Name</Label>
                <Input
                  id="new-label"
                  placeholder="VS Code Server"
                  value={newLabel}
                  onChange={(e) => setNewLabel(e.target.value)}
                />
              </div>
              <div className="flex items-end">
                <Button onClick={handleAddPort}>Add</Button>
              </div>
            </div>

            {error && (
              <p className="text-sm text-destructive" role="alert">
                {error}
              </p>
            )}

            <div className="flex gap-2">
              <Button onClick={handleSave} disabled={saving}>
                {saving ? 'Saving...' : 'Save Changes'}
              </Button>
              <Button variant="outline" onClick={handleCancel}>
                Cancel
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* Access Any Port Section */}
      {!isEditing && workspace.status === 'running' && (
        <div className="space-y-3">
          <h3 className="text-lg font-semibold">Access Any Port</h3>
          <p className="text-sm text-muted-foreground">
            Enter any port number to access it directly. All ports are
            accessible dynamically.
          </p>
          <div className="flex gap-2 max-w-md">
            <Input
              placeholder="Port number (e.g., 3000)"
              value={anyPort}
              onChange={(e) => setAnyPort(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleOpenAnyPort()}
            />
            <Button onClick={handleOpenAnyPort}>Open</Button>
          </div>
          {error && !isEditing && (
            <p className="text-sm text-destructive" role="alert">
              {error}
            </p>
          )}
        </div>
      )}

      {/* Status Warning */}
      {workspace.status !== 'running' && (
        <Alert>
          <AlertDescription>
            Port forwarding is only available when the workspace is running.
            Current status: <strong>{workspace.status}</strong>
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}
