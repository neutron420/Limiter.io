"use client"

import * as React from "react"
import { useProject } from "@/lib/project-context"
import { api } from "@/lib/api"
import { Spinner } from "@/components/dashboard/kit"
import { StatsGrid } from "@/components/dashboard/analytics/StatsGrid"
import { RequestChart } from "@/components/dashboard/analytics/RequestChart"
import { RatioChart } from "@/components/dashboard/analytics/RatioChart"
import { RoutesBreakdown } from "@/components/dashboard/analytics/RoutesBreakdown"
import { LatencyPercentiles } from "@/components/dashboard/analytics/LatencyPercentiles"
import { IpCountryBreakdown } from "@/components/dashboard/analytics/IpCountryBreakdown"
import { SecurityAnalytics } from "@/components/dashboard/analytics/SecurityAnalytics"
import { ApiKeyBreakdown } from "@/components/dashboard/analytics/ApiKeyBreakdown"
import { AuditLogsExplorer } from "@/components/dashboard/analytics/AuditLogsExplorer"

interface TimeSeriesItem {
  time: string
  allowed: number
  blocked: number
}

interface StatsData {
  total_requests?: number
  allowed_requests?: number
  blocked_requests?: number
  avg_latency_ms?: number
}

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

export default function AnalyticsPage() {
  const { current } = useProject()
  const [duration, setDuration] = React.useState("24h")
  const [bucket, setBucket] = React.useState("hour")
  const [stats, setStats] = React.useState<StatsData>({})
  const [timeseries, setTimeseries] = React.useState<TimeSeriesItem[]>([])
  const [logs, setLogs] = React.useState<LogItem[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)
  const [activeTab, setActiveTab] = React.useState("overview")

  const loadData = React.useCallback(async () => {
    if (!current) return
    setLoading(true)
    setError(null)
    try {
      // Fetch stats
      const s = await api.get<StatsData>(
        `/projects/${current.id}/analytics/stats?duration=${duration}`
      )
      setStats(s ?? {})

      // Fetch timeseries
      const ts = await api.get<TimeSeriesItem[]>(
        `/projects/${current.id}/analytics/timeseries?duration=${duration}&bucket=${bucket}`
      )
      setTimeseries(ts ?? [])

      // Fetch logs
      const rawLogs = await api.get<LogItem[]>(
        `/projects/${current.id}/analytics/logs?limit=250`
      )
      setLogs(rawLogs ?? [])
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load telemetry")
    } finally {
      setLoading(false)
    }
  }, [current, duration, bucket])

  React.useEffect(() => {
    loadData()
  }, [loadData])

  // Automatically adjust bucket when duration changes
  const handleDurationChange = (d: string) => {
    setDuration(d)
    if (d === "24h") setBucket("hour")
    else if (d === "7d") setBucket("hour")
    else if (d === "30d") setBucket("day")
  }

  if (!current) {
    return (
      <div className="flex h-full items-center justify-center p-6 text-center">
        <p className="text-xs uppercase tracking-widest text-muted-foreground animate-pulse">
          Select or create a project to see analytics...
        </p>
      </div>
    )
  }

  const allowedCount = stats.allowed_requests ?? 0
  const blockedCount = stats.blocked_requests ?? 0
  const totalCount = stats.total_requests ?? 0
  const avgLatency = stats.avg_latency_ms ?? 0

  return (
    <div className="flex flex-col gap-6 font-mono">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-lg font-bold uppercase tracking-widest">Analytics Dashboard</h1>
          <p className="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">
            Telemetry metrics and logs for project {current.name}
          </p>
        </div>

        {/* Start stark selection buttons */}
        <div className="flex items-center gap-2 border-2 border-foreground bg-background p-1 select-none w-fit">
          {["24h", "7d", "30d"].map((d) => (
            <button
              key={d}
              onClick={() => handleDurationChange(d)}
              className={`px-3 py-1 text-[10px] font-bold uppercase transition-all cursor-pointer ${
                duration === d
                  ? "bg-[#ea580c] text-white border-2 border-foreground shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] translate-x-[-1px] translate-y-[-1px]"
                  : "hover:bg-muted/10 text-foreground border-2 border-transparent"
              }`}
            >
              {d === "24h" ? "24 HOURS" : d === "7d" ? "7 DAYS" : "30 DAYS"}
            </button>
          ))}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex flex-wrap items-center gap-2 border-b-2 border-foreground pb-2 select-none">
        {[
          { id: "overview", label: "Overview" },
          { id: "routes", label: "Routes & Credentials" },
          { id: "security", label: "Security & Rules" },
          { id: "geolocation", label: "IPs & Countries" },
          { id: "logs", label: "Audit Log Explorer" },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-1.5 text-[10px] font-bold uppercase transition-all cursor-pointer border-2 ${
              activeTab === tab.id
                ? "bg-foreground text-background border-foreground shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] translate-x-[-1px] translate-y-[-1px]"
                : "hover:bg-muted/10 text-foreground border-transparent"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {loading && timeseries.length === 0 ? (
        <div className="py-20 flex justify-center">
          <Spinner label="FETCHING AUDIT TRAIL TELEMETRY" />
        </div>
      ) : error ? (
        <div className="rounded-none border-2 border-danger bg-danger/10 p-4 text-xs font-bold text-danger uppercase">
          Telemetry Fetch Error: {error}
        </div>
      ) : (
        <>
          {activeTab === "overview" && (
            <>
              {/* Key Stat Cards Grid Component */}
              <StatsGrid
                total={totalCount}
                allowed={allowedCount}
                blocked={blockedCount}
                avgLatency={avgLatency}
              />

              {/* Primary Charts Layout (Requests Over Time & Ratio Pie) */}
              <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                <RequestChart timeseries={timeseries} bucket={bucket} />
                <RatioChart allowed={allowedCount} blocked={blockedCount} />
              </div>
            </>
          )}

          {activeTab === "routes" && (
            <div className="flex flex-col gap-6">
              <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
                <RoutesBreakdown logs={logs} />
                <LatencyPercentiles logs={logs} />
              </div>
              <ApiKeyBreakdown logs={logs} />
            </div>
          )}

          {activeTab === "security" && (
            <SecurityAnalytics logs={logs} />
          )}

          {activeTab === "geolocation" && (
            <IpCountryBreakdown logs={logs} />
          )}

          {activeTab === "logs" && (
            <AuditLogsExplorer logs={logs} />
          )}
        </>
      )}
    </div>
  )
}

