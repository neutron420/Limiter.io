"use client"

import * as React from "react"
import { BarChart3, Clock, HelpCircle, Shield, ShieldAlert, Sparkles } from "lucide-react"
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  PieChart,
  Pie,
  Cell,
} from "recharts"
import { useProject } from "@/lib/project-context"
import { api } from "@/lib/api"
import {
  Panel,
  PanelHeader,
  StatCard,
  Label,
  Spinner,
} from "@/components/dashboard/kit"

interface TimeSeriesItem {
  time: string
  allowed: number
  blocked: number
}

interface StatsData {
  total_requests?: number
  allowed_requests?: number
  blocked_requests?: number
  average_latency_ms?: number
}

export default function AnalyticsPage() {
  const { current } = useProject()
  const [duration, setDuration] = React.useState("24h")
  const [bucket, setBucket] = React.useState("hour")
  const [stats, setStats] = React.useState<StatsData>({})
  const [timeseries, setTimeseries] = React.useState<TimeSeriesItem[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)

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
  const blockedRate = totalCount > 0 ? ((blockedCount / totalCount) * 100).toFixed(1) : "0.0"

  const pieData = [
    { name: "ALLOWED", value: allowedCount, color: "#22c55e" },
    { name: "BLOCKED", value: blockedCount, color: "#ef4444" },
  ]

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
          {/* Key Stat Cards */}
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard
              label="Total Requests"
              value={totalCount.toLocaleString()}
              icon={Sparkles}
              hint="All incoming requests processed"
            />
            <StatCard
              label="Allowed Requests"
              value={allowedCount.toLocaleString()}
              icon={Shield}
              hint="Requests within quota rules"
            />
            <StatCard
              label="Blocked Requests"
              value={blockedCount.toLocaleString()}
              icon={ShieldAlert}
              hint="Requests blocked by policy"
            />
            <StatCard
              label="Avg Response Latency"
              value={`${stats.average_latency_ms?.toFixed(2) ?? "0.00"} ms`}
              icon={Clock}
              hint="Telemetry validation latency"
            />
          </div>

          {/* Charts Layout */}
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
            {/* Timeseries Area Chart */}
            <Panel className="lg:col-span-2">
              <PanelHeader title="Requests Over Time" icon={BarChart3} />
              <div className="p-4 h-[300px] w-full text-xs">
                {timeseries.length === 0 ? (
                  <div className="flex h-full items-center justify-center text-muted-foreground uppercase text-[10px]">
                    No request history inside selected window.
                  </div>
                ) : (
                  <ResponsiveContainer width="100%" height="100%">
                    <AreaChart
                      data={timeseries}
                      margin={{ top: 10, right: 10, left: -20, bottom: 0 }}
                    >
                      <XAxis
                        dataKey="time"
                        tickFormatter={(t) => {
                          const date = new Date(t)
                          return bucket === "day"
                            ? date.toLocaleDateString(undefined, { month: "short", day: "numeric" })
                            : date.toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit" })
                        }}
                        stroke="#000"
                        style={{ fontSize: 9, fontFamily: "monospace" }}
                      />
                      <YAxis stroke="#000" style={{ fontSize: 9, fontFamily: "monospace" }} />
                      <Tooltip
                        contentStyle={{
                          backgroundColor: "#fff",
                          border: "2px solid #000",
                          borderRadius: 0,
                          fontFamily: "monospace",
                          fontSize: 10,
                        }}
                        labelFormatter={(label) => new Date(label).toLocaleString()}
                      />
                      <Area
                        type="monotone"
                        dataKey="allowed"
                        stackId="1"
                        stroke="#22c55e"
                        fill="#22c55e"
                        fillOpacity={0.3}
                        name="Allowed"
                      />
                      <Area
                        type="monotone"
                        dataKey="blocked"
                        stackId="1"
                        stroke="#ef4444"
                        fill="#ef4444"
                        fillOpacity={0.3}
                        name="Blocked"
                      />
                    </AreaChart>
                  </ResponsiveContainer>
                )}
              </div>
            </Panel>

            {/* Allowed vs Blocked ratio pie */}
            <Panel>
              <PanelHeader title="Allowed vs Blocked" icon={ShieldAlert} />
              <div className="p-4 flex flex-col items-center justify-center h-[300px] relative">
                {totalCount === 0 ? (
                  <div className="text-muted-foreground uppercase text-[10px]">
                    No ratio data to display.
                  </div>
                ) : (
                  <>
                    <div className="h-[180px] w-full">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={pieData}
                            cx="50%"
                            cy="50%"
                            innerRadius={45}
                            outerRadius={65}
                            paddingAngle={5}
                            dataKey="value"
                          >
                            {pieData.map((entry, index) => (
                              <Cell key={`cell-${index}`} fill={entry.color} stroke="#000" strokeWidth={2} />
                            ))}
                          </Pie>
                        </PieChart>
                      </ResponsiveContainer>
                    </div>

                    <div className="text-center mt-2 flex flex-col items-center">
                      <div className="text-2xl font-bold">{blockedRate}%</div>
                      <Label className="font-bold">BLOCKED RATE</Label>
                    </div>

                    {/* Legend */}
                    <div className="flex gap-4 mt-4 text-[10px] font-bold">
                      <div className="flex items-center gap-1.5">
                        <div className="w-2.5 h-2.5 bg-green-500 border border-black" />
                        <span>ALLOWED ({allowedCount})</span>
                      </div>
                      <div className="flex items-center gap-1.5">
                        <div className="w-2.5 h-2.5 bg-red-500 border border-black" />
                        <span>BLOCKED ({blockedCount})</span>
                      </div>
                    </div>
                  </>
                )}
              </div>
            </Panel>
          </div>
        </>
      )}
    </div>
  )
}
