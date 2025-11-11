import { useEffect, useState } from 'react'
import { useParams, useSearchParams, useNavigate, Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { PortsTab } from '@/components/workspace/PortsTab'
import { ConfigTab } from '@/components/workspace/ConfigTab'
import { workspaceApi } from '@/api/workspaces'
import type { Workspace } from '@/api/types'

export function WorkspaceDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()

  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Get initial tab from URL query parameter
  const initialTab = searchParams.get('tab') || 'terminal'

  const fetchWorkspace = async () => {
    if (!id) return

    try {
      setLoading(true)
      const { data } = await workspaceApi.get(id)
      setWorkspace(data)
      setError('')
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load workspace')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchWorkspace()

    // Auto-refresh every 5 seconds
    const interval = setInterval(fetchWorkspace, 5000)
    return () => clearInterval(interval)
  }, [id])

  if (loading && !workspace) {
    return (
      <div className="container mx-auto py-6">
        <p className="text-center text-muted-foreground">Loading workspace...</p>
      </div>
    )
  }

  if (error && !workspace) {
    return (
      <div className="container mx-auto py-6 space-y-4">
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
        <Button onClick={() => navigate('/')}>Back to Workspaces</Button>
      </div>
    )
  }

  if (!workspace) {
    return (
      <div className="container mx-auto py-6 space-y-4">
        <Alert variant="destructive">
          <AlertDescription>Workspace not found</AlertDescription>
        </Alert>
        <Button onClick={() => navigate('/')}>Back to Workspaces</Button>
      </div>
    )
  }

  // Status configuration
  const statusConfig = {
    creating: { color: 'default' as const, label: 'Creating...' },
    running: { color: 'default' as const, label: 'Running' },
    error: { color: 'destructive' as const, label: 'Error' },
    failed: { color: 'destructive' as const, label: 'Failed' },
  }

  const config = statusConfig[workspace.status]

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="space-y-4">
        <Link to="/">
          <Button variant="outline" size="sm">
            ← Back to Workspaces
          </Button>
        </Link>

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <h1 className="text-3xl font-bold">{workspace.name}</h1>
            <Badge variant={config.color}>{config.label}</Badge>
          </div>
        </div>
      </div>

      {/* Error Alert */}
      {workspace.error && (
        <Alert variant="destructive">
          <AlertDescription>{workspace.error}</AlertDescription>
        </Alert>
      )}

      {/* Tabs */}
      <Tabs defaultValue={initialTab} className="w-full">
        <TabsList>
          <TabsTrigger value="terminal">Terminal</TabsTrigger>
          <TabsTrigger value="ports">Ports</TabsTrigger>
          <TabsTrigger value="config">Config</TabsTrigger>
        </TabsList>

        {/* Terminal Tab - Placeholder for Module 6 */}
        <TabsContent value="terminal" className="mt-6">
          <div className="border rounded-lg p-12 text-center space-y-3">
            <h3 className="text-lg font-semibold">Terminal</h3>
            <p className="text-muted-foreground">
              Terminal integration will be implemented in Module 6
            </p>
            {workspace.status !== 'running' && (
              <Alert>
                <AlertDescription>
                  Terminal is only available when the workspace is running.
                </AlertDescription>
              </Alert>
            )}
          </div>
        </TabsContent>

        {/* Ports Tab */}
        <TabsContent value="ports" className="mt-6">
          <PortsTab workspace={workspace} onUpdate={fetchWorkspace} />
        </TabsContent>

        {/* Config Tab */}
        <TabsContent value="config" className="mt-6">
          <ConfigTab workspace={workspace} onUpdate={fetchWorkspace} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
