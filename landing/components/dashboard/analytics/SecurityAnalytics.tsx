"use client"

import * as React from "react"
import { ShieldAlert, AlertTriangle } from "lucide-react"
import { ResponsiveContainer, PieChart, Pie, Cell, Tooltip, BarChart, Bar, XAxis, YAxis } from "recharts"
import { Panel, PanelHeader } from "@/components/dashboard/kit"

interface LogItem {
  decision: string
  blocked_reason?: string
  route: string
}

interface SecurityAnalyticsProps {
  logs: LogItem[]
}

const COLORS = ["#ef4444", "#f97316", "#eab308", "#a855f7", "#3b82f6"]

export function SecurityAnalytics({ logs }: SecurityAnalyticsProps) {
  // Aggregate Blocked Reasons
  const reasonData = React.useMemo(() => {
    const counts: Record<string, number> = {}
    logs.forEach((log) => {
      if (log.decision !== "allowed") {
        let reason = log.blocked_reason || "unknown_block"
        // Format names nicely
        if (reason === "rate_limit_exceeded") reason = "Rate Limit Exceeded"
        else if (reason === "captcha_failed") reason = "CAPTCHA Failed"
        else if (reason === "invalid_api_key") reason = "Invalid API Key"
        else if (reason === "rule_match_blocked") reason = "Rule Block Blocked"
        
        counts[reason] = (counts[reason] || 0) + 1
      }
    })

    return Object.entries(counts).map(([name, value]) => ({
      name,
      value,
    }))
  }, [logs])

  // Aggregate Rules/Routes Triggering Blocks
  const ruleHits = React.useMemo(() => {
    const counts: Record<string, number> = {}
    logs.forEach((log) => {
      if (log.decision !== "allowed") {
        const route = log.route || "Unknown Route"
        counts[route] = (counts[route] || 0) + 1
      }
    })

    return Object.entries(counts)
      .map(([route, count]) => ({
        route,
        count,
      }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 5)
  }, [logs])

  const totalBlocked = React.useMemo(() => {
    return logs.filter((l) => l.decision !== "allowed").length
  }, [logs])

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
      {/* Blocked Reasons Pie Chart */}
      <Panel className="lg:col-span-1">
        <PanelHeader title="Blocked Reason Distribution" icon={ShieldAlert} />
        <div className="p-4 h-[300px] flex flex-col justify-between">
          {reasonData.length === 0 ? (
            <div className="flex h-full items-center justify-center text-muted-foreground uppercase text-[10px] text-center">
              No blocked requests recorded in audit logs.
            </div>
          ) : (
            <>
              <div className="h-[200px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={reasonData}
                      cx="50%"
                      cy="50%"
                      innerRadius={40}
                      outerRadius={70}
                      paddingAngle={4}
                      dataKey="value"
                    >
                      {reasonData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} stroke="#000" strokeWidth={2} />
                      ))}
                    </Pie>
                    <Tooltip
                      contentStyle={{
                        backgroundColor: "#fff",
                        border: "2px solid #000",
                        borderRadius: 0,
                        fontFamily: "monospace",
                        fontSize: 10,
                      }}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <div className="flex flex-wrap gap-2 text-[9px] font-bold uppercase tracking-wider justify-center">
                {reasonData.map((item, index) => (
                  <div key={item.name} className="flex items-center gap-1">
                    <span
                      className="inline-block w-2.5 h-2.5 border border-foreground"
                      style={{ backgroundColor: COLORS[index % COLORS.length] }}
                    />
                    <span>{item.name} ({item.value})</span>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </Panel>

      {/* Top Blocked Routes/Rules */}
      <Panel className="lg:col-span-2">
        <PanelHeader title="Top Triggered Security Restrictions" icon={AlertTriangle} />
        <div className="p-4 h-[300px]">
          {ruleHits.length === 0 ? (
            <div className="flex h-full items-center justify-center text-muted-foreground uppercase text-[10px] text-center">
              No rule hits triggered yet.
            </div>
          ) : (
            <div className="h-full flex flex-col justify-between">
              <div className="h-[200px] w-full text-xs">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart
                    data={ruleHits}
                    layout="vertical"
                    margin={{ top: 5, right: 10, left: 30, bottom: 5 }}
                  >
                    <XAxis type="number" stroke="#000" style={{ fontSize: 9, fontFamily: "monospace" }} />
                    <YAxis
                      dataKey="route"
                      type="category"
                      stroke="#000"
                      style={{ fontSize: 9, fontFamily: "monospace" }}
                      width={80}
                    />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: "#fff",
                        border: "2px solid #000",
                        borderRadius: 0,
                        fontFamily: "monospace",
                        fontSize: 10,
                      }}
                    />
                    <Bar dataKey="count" fill="#ef4444" stroke="#000" strokeWidth={2} name="Blocks Triggered" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
              <div className="text-[10px] uppercase text-muted-foreground font-bold border-t border-foreground/10 pt-2">
                A total of <span className="text-red-500 font-extrabold">{totalBlocked}</span> attempts were successfully rate-limited or blocked.
              </div>
            </div>
          )}
        </div>
      </Panel>
    </div>
  )
}
