import { useEffect, useRef, useState, useCallback } from 'react'
import { useAtom } from 'jotai'
import { tokenAtom } from '@/stores/auth'

export type WebSocketStatus = 'connecting' | 'connected' | 'disconnected'

export interface UseWebSocketReturn {
  ws: WebSocket | null
  status: WebSocketStatus
  reconnect: () => void
}

export function useWebSocket(url: string): UseWebSocketReturn {
  const [token] = useAtom(tokenAtom)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<number | null>(null)
  const [status, setStatus] = useState<WebSocketStatus>('disconnected')
  const urlRef = useRef(url)

  // Update url ref when url changes
  useEffect(() => {
    urlRef.current = url
  }, [url])

  const connect = useCallback(() => {
    if (!token) {
      console.warn('Cannot connect WebSocket: token is missing')
      return
    }

    // Clear any pending reconnection
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }

    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close()
    }

    setStatus('connecting')

    // Construct WebSocket URL with token as query parameter
    // Support both ws:// and wss:// protocols
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}${urlRef.current}?token=${token}`

    try {
      const ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        console.log('WebSocket connected')
        setStatus('connected')
      }

      ws.onclose = () => {
        console.log('WebSocket disconnected')
        setStatus('disconnected')

        // Auto-reconnect after 3 seconds
        reconnectTimeoutRef.current = setTimeout(() => {
          console.log('Attempting to reconnect...')
          connect()
        }, 3000)
      }

      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        setStatus('disconnected')
      }

      wsRef.current = ws
    } catch (error) {
      console.error('Failed to create WebSocket:', error)
      setStatus('disconnected')
    }
  }, [token])

  const reconnect = useCallback(() => {
    console.log('Manual reconnection triggered')
    connect()
  }, [connect])

  // Initial connection
  useEffect(() => {
    connect()

    // Cleanup on unmount
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [connect])

  return {
    ws: wsRef.current,
    status,
    reconnect,
  }
}
