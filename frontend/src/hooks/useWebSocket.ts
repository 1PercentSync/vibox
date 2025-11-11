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
  const reconnectAttemptsRef = useRef<number>(0)
  const [status, setStatus] = useState<WebSocketStatus>('disconnected')
  const urlRef = useRef(url)

  const MAX_RECONNECT_ATTEMPTS = 10
  const INITIAL_RECONNECT_DELAY = 1000

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
        // Reset reconnect attempts on successful connection
        reconnectAttemptsRef.current = 0
      }

      ws.onclose = () => {
        console.log('WebSocket disconnected')
        setStatus('disconnected')

        // Auto-reconnect with exponential backoff if under max attempts
        if (reconnectAttemptsRef.current < MAX_RECONNECT_ATTEMPTS) {
          const delay = Math.min(
            INITIAL_RECONNECT_DELAY * Math.pow(2, reconnectAttemptsRef.current),
            30000 // Max 30 seconds
          )
          reconnectAttemptsRef.current++

          console.log(`Attempting to reconnect (${reconnectAttemptsRef.current}/${MAX_RECONNECT_ATTEMPTS}) in ${delay}ms...`)
          reconnectTimeoutRef.current = setTimeout(() => {
            connect()
          }, delay)
        } else {
          console.warn('Max reconnection attempts reached. Please reconnect manually.')
        }
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
    // Reset attempts on manual reconnect
    reconnectAttemptsRef.current = 0
    connect()
  }, [connect])

  // Initial connection
  useEffect(() => {
    connect()

    // Don't cleanup WebSocket on unmount - keep connection alive
    // This allows terminal sessions to persist across page navigation
    return () => {
      console.log('useWebSocket unmounting, keeping connection alive')
      // Clear reconnect timeout but keep connection
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      // Don't close WebSocket - it will stay connected
    }
  }, [connect])

  return {
    ws: wsRef.current,
    status,
    reconnect,
  }
}
