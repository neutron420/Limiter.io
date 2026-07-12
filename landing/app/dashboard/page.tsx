"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { ColumnDef } from "@tanstack/react-table"
import { Activity, Check, ShieldAlert, Zap, RefreshCw, Boxes, Ban } from "lucide-react"

import { DataTable } from "@/components/data-table"
import {
  Panel,
  PanelHeader,
  StatCard,
  BrutalButton,
  Spinner,
  EmptyState,
  StatusBadge,
  InlineError,
  Label,
} from "@/components/dashboard/kit"
import { api } from "@/lib/api"
import { useProject } from "@/lib/project-context"
import { useAnalyticsWS } from "@/hooks/use-analytics-ws"
import type { AnalyticsLog, Stats } from "@/lib/types"

const DURATIONS = ["24h", "7d", "30d"] as const
type Duration = (typeof DURATIONS)[number]

const PAGE_SIZE = 50

function fmtTime(ts: string) {
  const d = new Date(ts)
  return isNaN(d.getTime()) ? "—" : d.toLocaleTimeString()
}

const columns: ColumnDef<AnalyticsLog>[] = [
  {
    accessorKey: "timestamp",
    header: "Time",
    cell: ({ row }) => <span className="text-muted-foreground">{fmtTime(row.original.timestamp)}</span>,
  },
  {
    accessorKey: "route",
    header: "Path",
    cell: ({ row }) => <span className="font-bold text-[#ea580c]">{row.original.route}</span>,
  },
  {
    accessorKey: "decision",
    header: "Status",
    cell: ({ row }) => <StatusBadge status={row.original.decision} />,
  },
  {
    accessorKey: "client_ip",
    header: "Client IP",
    cell: ({ row }) => <span>{row.original.client_ip}</span>,
  },
  {
    accessorKey: "status_code",
    header: "Code",
    cell: ({ row }) => <span className="tabular-nums">{row.original.status_code}</span>,
  },
  {
    accessorKey: "latency_ms",
    header: "Latency",
    cell: ({ row }) => <span className="tabular-nums">{row.original.latency_ms}ms</span>,
  },
]

export default function OverviewPage() {
  const router = useRouter()
  const { current, loading: projLoading, projects } = useProject()
  const projectId = current?.id ?? null

  const [duration, setDuration] = React.useState<Duration>("24h")
  const [stats, setStats] = React.useState<Stats | null>(null)
  const [statsLoading, setStatsLoading] = React.useState(false)
  const [logs, setLogs] = React.useState<AnalyticsLog[]>([])
  const [logsLoading, setLogsLoading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)

  const { events, status: wsStatus } = useAnalyticsWS(projectId)

  const loadStats = React.useCallback(async () => {
    if (!projectId) return
    setStatsLoading(true)
    setError(null)
    try {
      const s = await api.get<Stats>(`/projects/${projectId}/analytics/stats?duration=${duration}`)
      setStats(s)
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load stats")
    } finally {
      setStatsLoading(false)
    }
  }, [projectId, duration])

  const loadLogs = React.useCallback(async () => {
    if (!projectId) return
    setLogsLoading(true)
    try {
      const l = await api.get<AnalyticsLog[]>(
        `/projects/${projectId}/analytics/logs?limit=${PAGE_SIZE}&offset=0`,
      )
      setLogs(l ?? [])
    } catch {
      /* handled by stats error banner */
    } finally {
      setLogsLoading(false)
    }
  }, [projectId])

  React.useEffect(() => {
    loadStats()
  }, [loadStats])
  React.useEffect(() => {
    loadLogs()
  }, [loadLogs])

  // No project selected → guide the user.
  if (!projLoading && projects.length === 0) {
    return (
      <EmptyState
        icon={Boxes}
        title="No project selected"
        hint="Create a project to start collecting rate-limit analytics."
        action={
          <BrutalButton variant="primary" onClick={() => router.push("/dashboard/projects")}>
            Create a project
          </BrutalButton>
        }
      />
    )
  }

  if (projLoading || !current) return <Spinner label="LOADING PROJECT" />

  const topBlocked = stats?.top_blocked ?? []

  return (
    <div className="flex flex-col gap-6">
      {/* Header + duration switcher */}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-lg font-bold uppercase tracking-widest">Overview</h1>
          <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">
            {current.name} — traffic in the last {duration}
          </p>
        </div>
        <div className="flex items-center border-2 border-foreground">
          {DURATIONS.map((d) => (
            <button
              key={d}
              onClick={() => setDuration(d)}
              className={`px-3 py-1.5 text-xs font-bold uppercase tracking-wider transition-colors ${
                duration === d ? "bg-[#ea580c] text-white" : "hover:bg-muted/10"
              }`}
            >
              {d}
            </button>
          ))}
        </div>
      </div>

      <InlineError message={error} />

      {/* Stat cards */}
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Total Requests"
          value={statsLoading ? "…" : (stats?.total_requests ?? 0).toLocaleString()}
          hint="Evaluated by the gateway"
          icon={Activity}
        />
        <StatCard
          label="Allowed"
          value={statsLoading ? "…" : (stats?.allowed_requests ?? 0).toLocaleString()}
          hint="Passed active policies"
          icon={Check}
          iconClassName="text-green-500"
        />
        <StatCard
          label="Blocked"
          value={statsLoading ? "…" : (stats?.blocked_requests ?? 0).toLocaleString()}
          hint="Throttled by rate limits"
          icon={ShieldAlert}
          iconClassName="text-red-500"
        />
        <StatCard
          label="Avg Latency"
          value={statsLoading ? "…" : `${(stats?.avg_latency_ms ?? 0).toFixed(2)}ms`}
          hint="Mean evaluation time"
          icon={Zap}
        />
      </div>

      {/* Top blocked + live stream */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Panel className="lg:col-span-1">
          <PanelHeader title="Top Blocked Routes" icon={Ban} />
          <div className="flex flex-col divide-y-2 divide-foreground/10">
            {topBlocked.length === 0 ? (
              <p className="p-4 text-[11px] uppercase text-muted-foreground">No blocked traffic in range.</p>
            ) : (
              topBlocked.map((r) => (
                <div key={r.route} className="flex items-center justify-between gap-2 p-3">
                  <span className="truncate text-xs font-bold text-[#ea580c]">{r.route}</span>
                  <span className="shrink-0 border border-red-500/30 bg-red-500/10 px-2 py-0.5 text-[10px] font-bold tabular-nums text-red-500">
                    {r.count}
                  </span>
                </div>
              ))
            )}
          </div>
        </Panel>

        <Panel className="lg:col-span-2">
          <PanelHeader
            title="Live Stream"
            subtitle="Real-time gateway events via WebSocket"
            icon={Activity}
            action={
              <span className="flex items-center gap-1.5 text-[10px] font-bold uppercase tracking-wider">
                <span
                  className={`inline-block h-2 w-2 rounded-full ${
                    wsStatus === "open" ? "animate-pulse bg-green-500" : "bg-muted-foreground"
                  }`}
                />
                {wsStatus === "open" ? "LIVE" : wsStatus === "connecting" ? "CONNECTING" : "OFFLINE"}
              </span>
            }
          />
          <div className="max-h-[280px] overflow-auto">
            {events.length === 0 ? (
              <p className="p-4 text-[11px] uppercase text-muted-foreground">
                Waiting for live traffic… fire requests at{" "}
                <span className="text-foreground">/api/v1/gateway/*</span> to see events here.
              </p>
            ) : (
              <table className="w-full text-xs">
                <tbody>
                  {events.map((e) => (
                    <tr key={e.id} className="border-b border-foreground/10">
                      <td className="p-2 text-muted-foreground whitespace-nowrap">{fmtTime(e.timestamp)}</td>
                      <td className="p-2 font-bold text-[#ea580c] truncate max-w-[180px]">{e.route}</td>
                      <td className="p-2">
                        <StatusBadge status={e.decision} />
                      </td>
                      <td className="p-2 tabular-nums text-muted-foreground">{e.latency_ms}ms</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </Panel>
      </div>

      {/* Historical logs */}
      <Panel>
        <PanelHeader
          title="Request Logs"
          subtitle="Persisted analytics from PostgreSQL"
          icon={Activity}
          action={
            <BrutalButton variant="outline" icon={RefreshCw} loading={logsLoading} onClick={loadLogs}>
              Sync
            </BrutalButton>
          }
        />
        <div className="p-4 pt-0">
          {logsLoading && logs.length === 0 ? (
            <Spinner label="LOADING LOGS" />
          ) : (
            <DataTable columns={columns} data={logs} filterPlaceholder="Filter paths..." filterColumnKey="route" />
          )}
        </div>
      </Panel>
    </div>
  )
}
