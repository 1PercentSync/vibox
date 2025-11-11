import { useEffect, useRef, useState } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { WebglAddon } from '@xterm/addon-webgl'
import '@xterm/xterm/css/xterm.css'
import { useWebSocket } from '@/hooks/useWebSocket'
import { TerminalToolbar } from './TerminalToolbar'

interface TerminalProps {
  workspaceId: string
}

interface TerminalMessage {
  type: 'input' | 'output' | 'error' | 'resize'
  data?: string
  cols?: number
  rows?: number
}

export function Terminal({ workspaceId }: TerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<XTerm | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const [isFullscreen, setIsFullscreen] = useState(false)

  const wsUrl = `/ws/terminal/${workspaceId}`
  const { ws, status, reconnect } = useWebSocket(wsUrl)

  // Initialize terminal
  useEffect(() => {
    if (!terminalRef.current) return

    // Create terminal instance
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

    // Cleanup
    return () => {
      term.dispose()
    }
  }, [])

  // Handle WebSocket connection
  useEffect(() => {
    if (!ws || !termRef.current || status !== 'connected') return

    const term = termRef.current
    const fitAddon = fitAddonRef.current

    // Send user input to server
    const disposable = term.onData((data) => {
      const message: TerminalMessage = {
        type: 'input',
        data,
      }
      ws.send(JSON.stringify(message))
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
  }, [ws, status])

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
