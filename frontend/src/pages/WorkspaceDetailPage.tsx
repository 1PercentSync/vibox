import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Settings, Terminal as TerminalIcon, Network } from 'lucide-react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Terminal } from '@/components/terminal/Terminal'
import { PortsTab } from '@/components/workspace/PortsTab'
import { ConfigTab } from '@/components/workspace/ConfigTab'
import { workspaceApi } from '@/api/workspaces'
import type { Workspace } from '@/api/types'

export function WorkspaceDetailPage() {
  const { id } = useParams()
  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Fetch workspace details and auto-refresh
  useEffect(() => {
    const fetchWorkspace = async () => {
      if (!id) return

      try {
        if (loading) setLoading(true) // Only show loading on first fetch
        const { data } = await workspaceApi.get(id)
        setWorkspace(data)
        setError('')
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

  if (error && !workspace) {
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

        {/* Ports Tab - Using Module 5's PortsTab component */}
        <TabsContent value="ports" className="mt-6">
          <PortsTab workspace={workspace} onUpdate={async () => {
            try {
              const { data } = await workspaceApi.get(workspace.id)
              setWorkspace(data)
            } catch (err) {
              console.error('Failed to refresh workspace:', err)
            }
          }} />
        </TabsContent>

        {/* Config Tab - Using Module 5's ConfigTab component */}
        <TabsContent value="config" className="mt-6">
          <ConfigTab workspace={workspace} onUpdate={async () => {
            try {
              const { data } = await workspaceApi.get(workspace.id)
              setWorkspace(data)
            } catch (err) {
              console.error('Failed to refresh workspace:', err)
            }
          }} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
