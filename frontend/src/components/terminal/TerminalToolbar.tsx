import { RefreshCw, Maximize, Minimize, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import type { WebSocketStatus } from '@/hooks/useWebSocket'

interface TerminalToolbarProps {
  status: WebSocketStatus
  isFullscreen: boolean
  onReconnect: () => void
  onToggleFullscreen: () => void
  onClear: () => void
}

export function TerminalToolbar({
  status,
  isFullscreen,
  onReconnect,
  onToggleFullscreen,
  onClear,
}: TerminalToolbarProps) {
  const statusConfig = {
    connecting: { color: 'blue' as const, label: 'Connecting...' },
    connected: { color: 'green' as const, label: 'Connected' },
    disconnected: { color: 'red' as const, label: 'Disconnected' },
  }

  const config = statusConfig[status]

  return (
    <div className="flex items-center justify-between px-4 py-2 bg-gray-800 border-b border-gray-700 rounded-t-lg">
      <div className="flex items-center gap-3">
        <span className="text-sm font-medium text-gray-300">Terminal</span>
        <Badge variant={config.color as any}>{config.label}</Badge>
      </div>

      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          onClick={onReconnect}
          disabled={status === 'connected'}
          title="Reconnect"
          className="text-gray-300 hover:text-white hover:bg-gray-700"
        >
          <RefreshCw className="h-4 w-4" />
        </Button>

        <Button
          variant="ghost"
          size="sm"
          onClick={onClear}
          title="Clear terminal"
          className="text-gray-300 hover:text-white hover:bg-gray-700"
        >
          <Trash2 className="h-4 w-4" />
        </Button>

        <Button
          variant="ghost"
          size="sm"
          onClick={onToggleFullscreen}
          title={isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
          className="text-gray-300 hover:text-white hover:bg-gray-700"
        >
          {isFullscreen ? (
            <Minimize className="h-4 w-4" />
          ) : (
            <Maximize className="h-4 w-4" />
          )}
        </Button>
      </div>
    </div>
  )
}
