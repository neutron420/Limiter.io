"use client"

import * as React from "react"

import { API_BASE, tokens } from "@/lib/api"
import type { AnalyticsLog } from "@/lib/types"

type Status = "connecting" | "open" | "closed"

function wsUrl(projectId: string, token: string): string {
  // Derive ws(s):// from the configured http(s) API base — no hardcoded host.
  const base = API_BASE.replace(/^http/, "ws")
  return `${base}/projects/${projectId}/ws?token=${encodeURIComponent(token)}`
}

/**
 * Subscribes to the live analytics WebSocket for a project and keeps a rolling
 * buffer of the most recent events. Reconnects with backoff on drop.
 */
export function useAnalyticsWS(projectId: string | null, max = 100) {
  const [events, setEvents] = React.useState<AnalyticsLog[]>([])
  const [status, setStatus] = React.useState<Status>("closed")
  const wsRef = React.useRef<WebSocket | null>(null)
  const retryRef = React.useRef(0)
  const closedRef = React.useRef(false)

  React.useEffect(() => {
    if (!projectId) return
    closedRef.current = false
    let reconnectTimer: ReturnType<typeof setTimeout>

    const connect = () => {
      const token = tokens.access()
      if (!token) return
      setStatus("connecting")
      let ws: WebSocket
      try {
        ws = new WebSocket(wsUrl(projectId, token))
      } catch {
        scheduleReconnect()
        return
      }
      wsRef.current = ws

      ws.onopen = () => {
        retryRef.current = 0
        setStatus("open")
      }

      ws.onmessage = (e) => {
        // The server may batch multiple JSON records separated by newlines.
        const chunks = String(e.data).split("\n").filter(Boolean)
        const parsed: AnalyticsLog[] = []
        for (const chunk of chunks) {
          try {
            parsed.push(JSON.parse(chunk) as AnalyticsLog)
          } catch {
            /* ignore malformed frame */
          }
        }
        if (parsed.length) {
          setEvents((prev) => [...parsed.reverse(), ...prev].slice(0, max))
        }
      }

      ws.onclose = () => {
        setStatus("closed")
        if (!closedRef.current) scheduleReconnect()
      }

      ws.onerror = () => {
        ws.close()
      }
    }

    const scheduleReconnect = () => {
      const delay = Math.min(1000 * 2 ** retryRef.current, 15000)
      retryRef.current += 1
      reconnectTimer = setTimeout(connect, delay)
    }

    connect()

    return () => {
      closedRef.current = true
      clearTimeout(reconnectTimer)
      wsRef.current?.close()
      wsRef.current = null
    }
  }, [projectId, max])

  const clear = React.useCallback(() => setEvents([]), [])

  return { events, status, clear }
}
