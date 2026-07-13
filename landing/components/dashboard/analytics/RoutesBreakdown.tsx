"use client"

import * as React from "react"
import { Panel, PanelHeader } from "@/components/dashboard/kit"
import { Compass } from "lucide-react"

interface LogItem {
  id: string
  project_id: string
  api_key_id: string
  request_id: string
  client_ip: string
  route: string
  status_code: number
  latency_ms: number
  decision: string
  blocked_reason?: string
  timestamp: string
}

interface RoutesBreakdownProps {
  logs: LogItem[]
}

export function RoutesBreakdown({ logs }: RoutesBreakdownProps) {
  const breakdown = React.useMemo(() => {
    const counts: Record<string, { total: number; blocked: number }> = {}
    logs.forEach((log) => {
      const cleanPath = log.route.replace(/^\/api\/v1\/gateway/, "") || "/"
      if (!counts[cleanPath]) {
        counts[cleanPath] = { total: 0, blocked: 0 }
      }
      counts[cleanPath].total += 1
      if (log.decision === "blocked") {
        counts[cleanPath].blocked += 1
      }
    })

    return Object.entries(counts)
      .map(([path, data]) => ({
        path,
        total: data.total,
        blocked: data.blocked,
        blockedRate: data.total > 0 ? (data.blocked / data.total) * 100 : 0,
      }))
      .sort((a, b) => b.total - a.total)
      .slice(0, 5)
  }, [logs])

  return (
    <Panel>
      <PanelHeader title="Top API Endpoints" icon={Compass} />
      <div className="p-4 flex flex-col gap-4 min-h-[300px]">
        {breakdown.length === 0 ? (
          <div className="flex h-full flex-1 items-center justify-center text-muted-foreground uppercase text-[10px] my-auto">
            No endpoint telemetry captured yet.
          </div>
        ) : (
          breakdown.map((item) => {
            const pct = logs.length > 0 ? (item.total / logs.length) * 100 : 0
            return (
              <div key={item.path} className="flex flex-col gap-1">
                <div className="flex items-center justify-between text-xs font-bold">
                  <span className="font-mono text-foreground truncate max-w-[70%]" title={item.path}>
                    {item.path}
                  </span>
                  <span className="text-[10px] text-muted-foreground tabular-nums">
                    {item.total} REQS ({item.blocked} BLOCKED)
                  </span>
                </div>
                {/* Brutalist progress bar */}
                <div className="w-full h-4 border-2 border-foreground bg-background rounded-none overflow-hidden relative flex">
                  <div
                    style={{ width: `${pct}%` }}
                    className="h-full bg-[#ea580c] border-r-2 border-foreground"
                  />
                  {item.blocked > 0 && (
                    <div
                      style={{ width: `${(item.blocked / logs.length) * 100}%` }}
                      className="h-full bg-red-500 border-r-2 border-foreground absolute left-0"
                    />
                  )}
                </div>
                <div className="flex justify-between text-[9px] uppercase tracking-wider text-muted-foreground font-bold">
                  <span>{pct.toFixed(0)}% of total traffic</span>
                  <span className={item.blocked > 0 ? "text-red-500" : ""}>
                    {item.blockedRate.toFixed(1)}% blocked rate
                  </span>
                </div>
              </div>
            )
          })
        )}
      </div>
    </Panel>
  )
}
