"use client"

import * as React from "react"
import { Activity, Check, ShieldAlert, Zap, LogOut, RefreshCw } from "lucide-react"
import { useRouter } from "next/navigation"
import { api } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { useProject } from "@/lib/project-context"
import { useAnalyticsWS } from "@/hooks/use-analytics-ws"
import type { AnalyticsLog, Stats } from "@/lib/types"
import { Spinner, InlineError, Panel, PanelHeader, StatCard, StatusBadge, BrutalButton } from "@/components/dashboard/kit"

export default function MemberDashboardPage() {
  const router = useRouter()
  const { user, ready, logout } = useAuth()
  const { projects, current, role, loading, select } = useProject()
  const id = current?.id ?? null
  const { events, status } = useAnalyticsWS(id)
  const [stats, setStats] = React.useState<Stats | null>(null)
  const [logs, setLogs] = React.useState<AnalyticsLog[]>([])
  const [error, setError] = React.useState<string | null>(null)
  const [syncing, setSyncing] = React.useState(false)

  React.useEffect(() => {
    if (ready && !user) router.replace("/login")
    if (ready && user && role && role !== "member") router.replace("/dashboard")
  }, [ready, user, role, router])

  const sync = React.useCallback(async () => {
    if (!id) return
    setSyncing(true)
    try {
      const [nextStats, nextLogs] = await Promise.all([
        api.get<Stats>("/projects/" + id + "/analytics/stats?duration=all"),
        api.get<AnalyticsLog[]>("/projects/" + id + "/analytics/logs?limit=50&offset=0"),
      ])
      setStats(nextStats)
      setLogs(nextLogs ?? [])
      setError(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load incoming requests")
    } finally {
      setSyncing(false)
    }
  }, [id])

  React.useEffect(() => { sync() }, [sync])

  if (!ready || !user || loading || (role && role !== "member")) {
    return <div className="flex min-h-screen items-center justify-center"><Spinner label="LOADING MEMBER SPACE" /></div>
  }

  return (
    <main className="min-h-screen bg-background p-4 font-mono md:p-8">
      <div className="mx-auto flex max-w-7xl flex-col gap-6">
        <header className="flex flex-wrap items-center justify-between gap-4 border-b-2 border-foreground pb-5">
          <div>
            <p className="text-[10px] font-bold uppercase tracking-[0.3em] text-[#ea580c]">Limiter.io / Member Space</p>
            <h1 className="mt-2 text-2xl font-bold uppercase tracking-widest">{current?.name ?? "No project"}</h1>
            <p className="mt-1 text-xs uppercase text-muted-foreground">Read-only incoming request analytics</p>
          </div>
          <div className="flex items-center gap-2">
            {projects.length > 1 && <select className="border-2 border-foreground bg-background px-3 py-2 text-xs font-bold uppercase" value={current?.id ?? ""} onChange={(e) => select(e.target.value)}>{projects.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}</select>}
            <BrutalButton variant="outline" icon={LogOut} onClick={logout}>Log out</BrutalButton>
          </div>
        </header>

        <div className="border-2 border-yellow-500 bg-yellow-500/10 p-3 text-xs font-bold uppercase text-yellow-700">You are viewing this workspace as a read-only member.</div>
        <InlineError message={error} />

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
          <StatCard label="Total Requests" value={(stats?.total_requests ?? 0).toLocaleString()} hint="Incoming gateway traffic" icon={Activity} />
          <StatCard label="Allowed" value={(stats?.allowed_requests ?? 0).toLocaleString()} hint="Requests accepted" icon={Check} iconClassName="text-green-500" />
          <StatCard label="Blocked" value={(stats?.blocked_requests ?? 0).toLocaleString()} hint="Requests throttled" icon={ShieldAlert} iconClassName="text-red-500" />
          <StatCard label="Avg Latency" value={(stats?.avg_latency_ms ?? 0).toFixed(2) + "ms"} hint="Mean evaluation time" icon={Zap} />
        </div>

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <Panel>
            <PanelHeader title="Live Incoming Requests" subtitle="Real-time gateway events" icon={Activity} action={<span className="text-[10px] font-bold uppercase">{status === "open" ? "LIVE" : status}</span>} />
            <div className="max-h-[360px] overflow-auto">{events.length === 0 ? <p className="p-4 text-xs uppercase text-muted-foreground">Waiting for incoming requests...</p> : events.map((e) => <div key={e.id} className="flex items-center justify-between gap-3 border-b-2 border-foreground/10 p-3 text-xs"><span className="truncate font-bold text-[#ea580c]">{e.route}</span><StatusBadge status={e.decision} /><span className="text-muted-foreground">{e.latency_ms}ms</span></div>)}</div>
          </Panel>
          <Panel>
            <PanelHeader title="Incoming Request Log" subtitle="Persisted analytics" icon={Activity} action={<BrutalButton variant="outline" icon={RefreshCw} loading={syncing} onClick={sync}>Sync</BrutalButton>} />
            <div className="max-h-[360px] overflow-auto">{logs.length === 0 ? <p className="p-4 text-xs uppercase text-muted-foreground">No incoming requests recorded.</p> : logs.map((log) => <div key={log.id} className="grid grid-cols-[1fr_auto_auto] items-center gap-3 border-b-2 border-foreground/10 p-3 text-xs"><span className="truncate font-bold text-[#ea580c]">{log.route}</span><StatusBadge status={log.decision} /><span className="text-muted-foreground">{log.status_code}</span></div>)}</div>
          </Panel>
        </div>
      </div>
    </main>
  )
}
