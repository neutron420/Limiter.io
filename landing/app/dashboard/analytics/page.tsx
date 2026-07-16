"use client"

import * as React from "react"
import { useProject } from "@/lib/project-context"
import { API_BASE, api, tokens } from "@/lib/api"
import { Download, Save, Eye, Search, AlertTriangle } from "lucide-react"

import {
  BrutalButton,
  Spinner,
  Panel,
  PanelHeader,
  Field,
  Modal,
  InlineError,
  StatusBadge,
  SubmitButton,
} from "@/components/dashboard/kit"
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

import type { SavedAnalyticsView, AnomalyDetectionConfig } from "@/lib/types"

export default function AnalyticsPage() {
  const { current } = useProject()
  const [duration, setDuration] = React.useState("all")
  const [bucket, setBucket] = React.useState("day")
  const [stats, setStats] = React.useState<StatsData>({})
  const [timeseries, setTimeseries] = React.useState<TimeSeriesItem[]>([])
  const [logs, setLogs] = React.useState<LogItem[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)
  const [activeTab, setActiveTab] = React.useState("overview")
  const [exporting, setExporting] = React.useState<"csv" | "json" | null>(null)

  // Saved views
  const [views, setViews] = React.useState<SavedAnalyticsView[]>([])
  const [showViewModal, setShowViewModal] = React.useState(false)
  const [viewName, setViewName] = React.useState("")
  const [savingView, setSavingView] = React.useState(false)

  // Anomaly detection
  const [anomalyConfig, setAnomalyConfig] = React.useState<AnomalyDetectionConfig | null>(null)
  const [anomalySensitivity, setAnomalySensitivity] = React.useState("2")
  const [anomalyEnabled, setAnomalyEnabled] = React.useState(false)
  const [anomalyLookback, setAnomalyLookback] = React.useState("60")
  const [anomalySpike, setAnomalySpike] = React.useState(true)
  const [anomalyDrop, setAnomalyDrop] = React.useState(true)
  const [anomalySlack, setAnomalySlack] = React.useState("")
  const [savingAnomaly, setSavingAnomaly] = React.useState(false)

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

      // Fetch saved views
      const v = await api.get<SavedAnalyticsView[]>(`/projects/${current.id}/analytics/views`).catch(() => [])
      setViews(v)

      // Fetch anomaly config
      const a = await api.get<AnomalyDetectionConfig>(`/projects/${current.id}/analytics/anomaly-config`).catch(() => null)
      setAnomalyConfig(a)
      if (a) {
        setAnomalyEnabled(a.enabled)
        setAnomalySensitivity(String(a.sensitivity))
        setAnomalyLookback(String(a.lookback_minutes))
        setAnomalySpike(a.alert_on_spike)
        setAnomalyDrop(a.alert_on_drop)
        setAnomalySlack(a.slack_webhook_url || "")
      }
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
    else if (d === "all") setBucket("day")
  }

  const downloadExport = async (format: "csv" | "json") => {
    if (!current) return
    setExporting(format)
    setError(null)
    try {
      const res = await fetch(
        `${API_BASE}/projects/${current.id}/analytics/export?format=${format}&limit=5000`,
        {
          headers: {
            Authorization: `Bearer ${tokens.access() ?? ""}`,
          },
        },
      )
      if (!res.ok) throw new Error(`Export failed with status ${res.status}`)

      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const link = document.createElement("a")
      const safeName = current.name.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "") || "project"
      link.href = url
      link.download = `${safeName}-analytics.${format}`
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(url)
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to export analytics")
    } finally {
      setExporting(null)
    }
  }

  const handleSaveView = async () => {
    if (!current) return
    setSavingView(true)
    try {
      await api.post(`/projects/${current.id}/analytics/views`, {
        name: viewName,
        config: JSON.stringify({ duration, bucket }),
        is_shared: false,
      })
      setViewName("")
      setShowViewModal(false)
      loadData()
    } catch {
      setError("Failed to save view")
    } finally {
      setSavingView(false)
    }
  }

  const handleSaveAnomaly = async () => {
    if (!current) return
    setSavingAnomaly(true)
    try {
      const data = await api.put<AnomalyDetectionConfig>(`/projects/${current.id}/analytics/anomaly-config`, {
        enabled: anomalyEnabled,
        sensitivity: parseFloat(anomalySensitivity) || 2,
        lookback_minutes: parseInt(anomalyLookback) || 60,
        alert_on_spike: anomalySpike,
        alert_on_drop: anomalyDrop,
        slack_webhook_url: anomalySlack,
      })
      setAnomalyConfig(data)
      setError(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save anomaly config")
    } finally {
      setSavingAnomaly(false)
    }
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

        <div className="flex flex-col gap-2 sm:items-end">
          <div className="flex flex-wrap items-center gap-2">
            <BrutalButton icon={Download} loading={exporting === "csv"} onClick={() => downloadExport("csv")}>
              CSV
            </BrutalButton>
            <BrutalButton icon={Download} loading={exporting === "json"} onClick={() => downloadExport("json")}>
              JSON
            </BrutalButton>
          </div>

          {/* Start stark selection buttons */}
          <div className="flex items-center gap-2 border-2 border-foreground bg-background p-1 select-none w-fit">
            {["24h", "7d", "30d", "all"].map((d) => (
              <button
                key={d}
                onClick={() => handleDurationChange(d)}
                className={`px-3 py-1 text-[10px] font-bold uppercase transition-all cursor-pointer ${
                  duration === d
                    ? "bg-[#ea580c] text-white border-2 border-foreground shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] translate-x-[-1px] translate-y-[-1px]"
                    : "hover:bg-muted/10 text-foreground border-2 border-transparent"
                }`}
              >
                {d === "24h" ? "24 HOURS" : d === "7d" ? "7 DAYS" : d === "30d" ? "30 DAYS" : "ALL TIME"}
              </button>
            ))}
          </div>
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
          { id: "views", label: "Views & Anomalies" },
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

          {activeTab === "views" && (
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-bold uppercase tracking-wider">Saved Views</h2>
                <BrutalButton onClick={() => setShowViewModal(true)} icon={Save} size="sm">SAVE CURRENT VIEW</BrutalButton>
              </div>
              {views.length === 0 ? (
                <p className="text-xs text-muted-foreground">No saved views yet</p>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                  {views.map((v) => (
                    <div key={v.id} className="rounded-base border-2 border-foreground p-3">
                      <p className="text-sm font-mono font-bold">{v.name}</p>
                      <p className="text-xs text-muted-foreground mt-1">
                        {v.is_shared && "Shared · "}
                        {new Date(v.created_at).toLocaleDateString()}
                      </p>
                    </div>
                  ))}
                </div>
              )}

              <Panel>
                <PanelHeader icon={AlertTriangle} title="Anomaly Detection" />
                <div className="space-y-4">
                  <label className="flex items-center justify-between rounded-base border-2 border-foreground p-3 cursor-pointer">
                    <div>
                      <p className="text-sm font-mono font-bold">Enable Anomaly Detection</p>
                      <p className="text-xs text-muted-foreground">Detect traffic anomalies automatically</p>
                    </div>
                    <input type="checkbox" checked={anomalyEnabled} onChange={(e) => setAnomalyEnabled(e.target.checked)} className="size-5 accent-foreground" />
                  </label>
                  <Field label="Sensitivity" sub="Number of standard deviations for threshold">
                    <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" type="number" step="0.1" value={anomalySensitivity} onChange={(e) => setAnomalySensitivity(e.target.value)} />
                  </Field>
                  <Field label="Lookback (minutes)" sub="How far back to analyze">
                    <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" type="number" value={anomalyLookback} onChange={(e) => setAnomalyLookback(e.target.value)} />
                  </Field>
                  <label className="flex items-center justify-between rounded-base border-2 border-foreground p-3 cursor-pointer">
                    <div>
                      <p className="text-sm font-mono font-bold">Alert on Spike</p>
                      <p className="text-xs text-muted-foreground">Alert when traffic spikes above threshold</p>
                    </div>
                    <input type="checkbox" checked={anomalySpike} onChange={(e) => setAnomalySpike(e.target.checked)} className="size-5 accent-foreground" />
                  </label>
                  <label className="flex items-center justify-between rounded-base border-2 border-foreground p-3 cursor-pointer">
                    <div>
                      <p className="text-sm font-mono font-bold">Alert on Drop</p>
                      <p className="text-xs text-muted-foreground">Alert when traffic drops below threshold</p>
                    </div>
                    <input type="checkbox" checked={anomalyDrop} onChange={(e) => setAnomalyDrop(e.target.checked)} className="size-5 accent-foreground" />
                  </label>
                  <Field label="Slack Webhook" sub="Optional Slack webhook for alerts">
                    <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" placeholder="https://hooks.slack.com/services/..." value={anomalySlack} onChange={(e) => setAnomalySlack(e.target.value)} />
                  </Field>
                  <SubmitButton onClick={handleSaveAnomaly} loading={savingAnomaly} icon={Save}>SAVE ANOMALY CONFIG</SubmitButton>
                </div>
              </Panel>
            </div>
          )}
        </>
      )}

      <Modal open={showViewModal} onClose={() => setShowViewModal(false)} title="Save Current View">
        <div className="space-y-4">
          <Field label="View Name">
            <input className="w-full rounded-base border-2 border-foreground bg-main px-3 py-2 font-mono text-sm" value={viewName} onChange={(e) => setViewName(e.target.value)} />
          </Field>
          <SubmitButton onClick={handleSaveView} loading={savingView} icon={Save}>SAVE</SubmitButton>
        </div>
      </Modal>
    </div>
  )
}

