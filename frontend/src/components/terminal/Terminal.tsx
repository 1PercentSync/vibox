import { useEffect, useRef, useState } from 'react'
import { useAtom, useAtomValue } from 'jotai'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { WebglAddon } from '@xterm/addon-webgl'
import '@xterm/xterm/css/xterm.css'
import { TerminalToolbar } from './TerminalToolbar'
import {
  terminalInstancesAtom,
  getTerminalInstance,
  setTerminalInstance,
  updateTerminalWebSocket,
} from '@/stores/terminals'
import { tokenAtom } from '@/stores/auth'

interface TerminalProps {
  workspaceId: string
}

interface TerminalMessage {
  type: 'input' | 'output' | 'error' | 'resize'
  data?: string
  cols?: number
  rows?: number
}

type WebSocketStatus = 'connecting' | 'connected' | 'disconnected'

export function Terminal({ workspaceId }: TerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<XTerm | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [status, setStatus] = useState<WebSocketStatus>('disconnected')
  const [terminalInstances, setTerminalInstances] = useAtom(terminalInstancesAtom)
  const token = useAtomValue(tokenAtom)

  // Initialize or restore terminal
  useEffect(() => {
    if (!terminalRef.current) return

    // Check if we have an existing terminal instance for this workspace
    const existingInstance = getTerminalInstance(terminalInstances, workspaceId)

    if (existingInstance) {
      // Reuse existing terminal
      console.log('Restoring existing terminal for workspace:', workspaceId)
      const { terminal, fitAddon, websocket } = existingInstance

      // Re-attach to DOM
      terminal.open(terminalRef.current)
      fitAddon.fit()

      termRef.current = terminal
      fitAddonRef.current = fitAddon
      wsRef.current = websocket

      // Update status based on WebSocket state
      if (websocket && websocket.readyState === WebSocket.OPEN) {
        setStatus('connected')
      } else if (websocket && websocket.readyState === WebSocket.CONNECTING) {
        setStatus('connecting')
      } else {
        setStatus('disconnected')
      }

      return // Don't dispose on unmount
    }

    // Create new terminal instance
    console.log('Creating new terminal for workspace:', workspaceId)
    const term = new XTerm({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#ffffff',
        selectionBackground: 'rgba(255, 255, 255, 0.3)',
        black: '#000000',
        red: '#cd3131',
        green: '#0dbc79',
        yellow: '#e5e510',
        blue: '#2472c8',
        magenta: '#bc3fbc',
        cyan: '#11a8cd',
        white: '#e5e5e5',
        brightBlack: '#666666',
        brightRed: '#f14c4c',
        brightGreen: '#23d18b',
        brightYellow: '#f5f543',
        brightBlue: '#3b8eea',
        brightMagenta: '#d670d6',
        brightCyan: '#29b8db',
        brightWhite: '#ffffff',
      },
      scrollback: 10000,
      convertEol: true,
    })

    // Load addons
    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()

    term.loadAddon(fitAddon)
    term.loadAddon(webLinksAddon)

    // Try to load WebGL addon for better performance
    try {
      const webglAddon = new WebglAddon()
      term.loadAddon(webglAddon)
    } catch (e) {
      // WebGL not supported, fallback to canvas
      console.warn('WebGL not supported, using canvas renderer:', e)
    }

    // Open terminal
    term.open(terminalRef.current)
    fitAddon.fit()

    termRef.current = term
    fitAddonRef.current = fitAddon

    // Store terminal instance for reuse
    setTerminalInstances((prev) =>
      setTerminalInstance(prev, workspaceId, term, fitAddon)
    )

    // Don't dispose on unmount - keep terminal alive for navigation
    return () => {
      // Terminal stays alive in memory
      console.log('Terminal component unmounting, keeping terminal alive:', workspaceId)
    }
  }, [workspaceId, terminalInstances, setTerminalInstances])

  // WebSocket connection management
  useEffect(() => {
    if (!token || !termRef.current) return

    const existingInstance = getTerminalInstance(terminalInstances, workspaceId)

    // If we already have a connected WebSocket, reuse it
    if (existingInstance?.websocket && existingInstance.websocket.readyState === WebSocket.OPEN) {
      console.log('Reusing existing WebSocket connection')
      wsRef.current = existingInstance.websocket
      setStatus('connected')
      return
    }

    // Create new WebSocket connection
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/ws/terminal/${workspaceId}?token=${token}`

    console.log('Creating new WebSocket connection')
    setStatus('connecting')

    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      console.log('WebSocket connected')
      setStatus('connected')
      wsRef.current = ws

      // Update terminal instance with WebSocket
      setTerminalInstances((prev) =>
        updateTerminalWebSocket(prev, workspaceId, ws)
      )
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected')
      setStatus('disconnected')
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      setStatus('disconnected')
    }

    wsRef.current = ws

    // Don't close WebSocket on unmount - keep it alive
    return () => {
      console.log('WebSocket effect cleanup, keeping connection alive')
    }
  }, [workspaceId, token, terminalInstances, setTerminalInstances])

  // Handle WebSocket data exchange and terminal events
  useEffect(() => {
    const ws = wsRef.current
    const term = termRef.current
    const fitAddon = fitAddonRef.current

    if (!ws || !term || status !== 'connected') return

    // Send user input to server
    const disposable = term.onData((data) => {
      const message: TerminalMessage = {
        type: 'input',
        data,
      }
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(message))
      }
    })

    // Receive output from server
    const handleMessage = (event: MessageEvent) => {
      try {
        const msg: TerminalMessage = JSON.parse(event.data)

        if (msg.type === 'output' && msg.data) {
          term.write(msg.data)
        } else if (msg.type === 'error' && msg.data) {
          term.write(`\r\n\x1b[31mError: ${msg.data}\x1b[0m\r\n`)
        }
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }

    ws.addEventListener('message', handleMessage)

    // Handle terminal resize
    const handleResize = () => {
      if (!fitAddon) return

      fitAddon.fit()

      const message: TerminalMessage = {
        type: 'resize',
        cols: term.cols,
        rows: term.rows,
      }

      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(message))
      }
    }

    // Send initial size
    handleResize()

    // Listen to window resize
    window.addEventListener('resize', handleResize)

    // Cleanup
    return () => {
      disposable.dispose()
      ws.removeEventListener('message', handleMessage)
      window.removeEventListener('resize', handleResize)
    }
  }, [status])

  // Reconnect function
  const reconnect = () => {
    // Close existing WebSocket if any
    if (wsRef.current) {
      wsRef.current.close()
    }

    // Remove from instances to force recreation
    setTerminalInstances((prev) => {
      const newInstances = new Map(prev)
      const instance = newInstances.get(workspaceId)
      if (instance) {
        newInstances.set(workspaceId, {
          ...instance,
          websocket: null,
        })
      }
      return newInstances
    })

    // Status will be updated by WebSocket effect
    setStatus('connecting')
  }

  // Handle fullscreen toggle
  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen)
    // Trigger resize after fullscreen change
    setTimeout(() => {
      fitAddonRef.current?.fit()
    }, 100)
  }

  // Clear terminal
  const clearTerminal = () => {
    termRef.current?.clear()
  }

  return (
    <div
      className={`flex flex-col ${
        isFullscreen ? 'fixed inset-0 z-50 bg-[#1e1e1e]' : 'h-full'
      }`}
    >
      <TerminalToolbar
        status={status}
        isFullscreen={isFullscreen}
        onReconnect={reconnect}
        onToggleFullscreen={toggleFullscreen}
        onClear={clearTerminal}
      />

      <div className="flex-1 overflow-hidden bg-[#1e1e1e] rounded-b-lg">
        {status !== 'connected' && (
          <div className="absolute inset-0 flex items-center justify-center bg-black/70 text-white z-10">
            <div className="text-center">
              <div className="text-lg mb-2">
                {status === 'connecting' ? 'Connecting...' : 'Disconnected'}
              </div>
              {status === 'disconnected' && (
                <button
                  onClick={reconnect}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm"
                >
                  Reconnect
                </button>
              )}
            </div>
          </div>
        )}
        <div ref={terminalRef} className="h-full w-full p-2" />
      </div>
    </div>
  )
}
