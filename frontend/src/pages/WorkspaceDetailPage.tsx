import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Settings, Terminal as TerminalIcon, Network } from 'lucide-react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Terminal } from '@/components/terminal/Terminal'
import { useEffect, useState } from 'react'
import { workspaceApi } from '@/api/workspaces'
import type { Workspace } from '@/api/types'

export function WorkspaceDetailPage() {
  const { id } = useParams()
  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Fetch workspace details
  useEffect(() => {
    const fetchWorkspace = async () => {
      if (!id) return

      try {
        setLoading(true)
        const { data } = await workspaceApi.get(id)
        setWorkspace(data)
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to load workspace')
      } finally {
        setLoading(false)
      }
    }

    fetchWorkspace()

    // Poll workspace status every 5 seconds
    const interval = setInterval(fetchWorkspace, 5000)
    return () => clearInterval(interval)
  }, [id])

  if (loading && !workspace) {
    return (
      <div className="container mx-auto py-6">
        <div className="text-center">Loading...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container mx-auto py-6">
        <div className="text-center text-red-500">{error}</div>
      </div>
    )
  }

  if (!workspace) {
    return (
      <div className="container mx-auto py-6">
        <div className="text-center">Workspace not found</div>
      </div>
    )
  }

  const statusConfig = {
    creating: { color: 'blue' as const, label: 'Creating...' },
    running: { color: 'green' as const, label: 'Running' },
    error: { color: 'orange' as const, label: 'Error' },
    failed: { color: 'red' as const, label: 'Failed' },
  }

  const config = statusConfig[workspace.status]

  // Check if terminal is available
  const canUseTerminal = workspace.status === 'running' ||
    (workspace.status === 'error' && workspace.container_id)

  return (
    <div className="container mx-auto py-6">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-center gap-4 mb-4">
          <Link to="/">
            <Button variant="ghost" size="sm">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Workspaces
            </Button>
          </Link>
        </div>

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <h1 className="text-3xl font-bold">{workspace.name}</h1>
            <Badge variant={config.color as any}>{config.label}</Badge>
          </div>

          <div className="text-sm text-muted-foreground">
            ID: {workspace.id}
          </div>
        </div>

        {workspace.error && (
          <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-lg text-red-800">
            <strong>Error:</strong> {workspace.error}
          </div>
        )}
      </div>

      {/* Tabs */}
      <Tabs defaultValue="terminal" className="w-full">
        <TabsList>
          <TabsTrigger value="terminal" disabled={!canUseTerminal}>
            <TerminalIcon className="h-4 w-4 mr-2" />
            Terminal
          </TabsTrigger>
          <TabsTrigger value="ports">
            <Network className="h-4 w-4 mr-2" />
            Ports
          </TabsTrigger>
          <TabsTrigger value="config">
            <Settings className="h-4 w-4 mr-2" />
            Config
          </TabsTrigger>
        </TabsList>

        {/* Terminal Tab */}
        <TabsContent value="terminal" className="h-[600px]">
          {canUseTerminal ? (
            <Terminal workspaceId={workspace.id} />
          ) : (
            <div className="flex items-center justify-center h-full border rounded-lg bg-gray-50">
              <div className="text-center text-gray-500">
                <TerminalIcon className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p className="text-lg font-medium">Terminal not available</p>
                <p className="text-sm mt-2">
                  {workspace.status === 'creating'
                    ? 'Workspace is being created...'
                    : 'Container is not running'}
                </p>
              </div>
            </div>
          )}
        </TabsContent>

        {/* Ports Tab */}
        <TabsContent value="ports" className="min-h-[400px]">
          <div className="border rounded-lg p-6">
            <h2 className="text-xl font-semibold mb-4">Port Management</h2>
            <p className="text-muted-foreground mb-4">
              Manage port forwarding and labels for this workspace.
            </p>

            {workspace.ports && Object.keys(workspace.ports).length > 0 ? (
              <div className="space-y-2">
                <h3 className="font-medium">Configured Ports:</h3>
                {Object.entries(workspace.ports).map(([port, label]) => (
                  <div
                    key={port}
                    className="flex items-center justify-between p-3 border rounded"
                  >
                    <div>
                      <span className="font-medium">{label}</span>
                      <span className="text-muted-foreground ml-2">:{port}</span>
                    </div>
                    {workspace.status === 'running' && (
                      <Button
                        size="sm"
                        onClick={() =>
                          window.open(`/forward/${workspace.id}/${port}/`, '_blank')
                        }
                      >
                        Open
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-muted-foreground italic">
                No port labels configured. You can still access any port dynamically.
              </p>
            )}

            <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
              <p className="text-sm text-blue-800">
                <strong>Note:</strong> All ports are accessible dynamically at{' '}
                <code className="bg-white px-1 py-0.5 rounded">
                  /forward/{'{workspace-id}'}/{'{port}'}
                </code>
              </p>
            </div>
          </div>
        </TabsContent>

        {/* Config Tab */}
        <TabsContent value="config" className="min-h-[400px]">
          <div className="border rounded-lg p-6">
            <h2 className="text-xl font-semibold mb-4">Configuration</h2>

            <div className="space-y-4">
              <div>
                <h3 className="text-sm font-medium text-muted-foreground">
                  Workspace ID
                </h3>
                <p className="text-sm mt-1">{workspace.id}</p>
              </div>

              <div>
                <h3 className="text-sm font-medium text-muted-foreground">Name</h3>
                <p className="text-sm mt-1">{workspace.name}</p>
              </div>

              <div>
                <h3 className="text-sm font-medium text-muted-foreground">Status</h3>
                <div className="mt-1">
                  <Badge variant={config.color as any}>{config.label}</Badge>
                </div>
              </div>

              {workspace.container_id && (
                <div>
                  <h3 className="text-sm font-medium text-muted-foreground">
                    Container ID
                  </h3>
                  <p className="text-sm mt-1 font-mono">{workspace.container_id}</p>
                </div>
              )}

              <div>
                <h3 className="text-sm font-medium text-muted-foreground">Image</h3>
                <p className="text-sm mt-1">{workspace.config.image}</p>
              </div>

              <div>
                <h3 className="text-sm font-medium text-muted-foreground">
                  Created At
                </h3>
                <p className="text-sm mt-1">
                  {new Date(workspace.created_at).toLocaleString()}
                </p>
              </div>

              {workspace.config.scripts && workspace.config.scripts.length > 0 && (
                <div>
                  <h3 className="text-sm font-medium text-muted-foreground mb-2">
                    Initialization Scripts
                  </h3>
                  <div className="space-y-3">
                    {workspace.config.scripts
                      .sort((a, b) => a.order - b.order)
                      .map((script, index) => (
                        <div key={index} className="border rounded p-3">
                          <div className="font-medium text-sm mb-2">
                            {script.order}. {script.name}
                          </div>
                          <pre className="text-xs bg-gray-50 p-3 rounded overflow-x-auto">
                            {script.content}
                          </pre>
                        </div>
                      ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
