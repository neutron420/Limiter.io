"use client"

import * as React from "react"
import { Key } from "lucide-react"
import { Panel, PanelHeader } from "@/components/dashboard/kit"

interface LogItem {
  api_key_id: string
  decision: string
  latency_ms: number
}

interface ApiKeyBreakdownProps {
  logs: LogItem[]
}

export function ApiKeyBreakdown({ logs }: ApiKeyBreakdownProps) {
  const stats = React.useMemo(() => {
    const keyData: Record<string, { total: number; allowed: number; blocked: number; sumLatency: number }> = {}
    
    logs.forEach((log) => {
      const keyId = log.api_key_id || "public / anonymous"
      if (!keyData[keyId]) {
        keyData[keyId] = { total: 0, allowed: 0, blocked: 0, sumLatency: 0 }
      }
      keyData[keyId].total++
      keyData[keyId].sumLatency += log.latency_ms || 0
      if (log.decision === "allowed") {
        keyData[keyId].allowed++
      } else {
        keyData[keyId].blocked++
      }
    })

    return Object.entries(keyData)
      .map(([keyId, d]) => ({
        keyId,
        total: d.total,
        allowed: d.allowed,
        blocked: d.blocked,
        avgLatency: d.total > 0 ? Math.round(d.sumLatency / d.total) : 0,
        successRate: d.total > 0 ? Math.round((d.allowed / d.total) * 100) : 0,
      }))
      .sort((a, b) => b.total - a.total)
  }, [logs])

  return (
    <Panel>
      <PanelHeader title="Request Volume by API Key Credentials" icon={Key} />
      <div className="p-4 overflow-x-auto">
        {stats.length === 0 ? (
          <p className="text-[10px] text-muted-foreground uppercase text-center py-8">
            No API key requests logged in telemetry window.
          </p>
        ) : (
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="border-b-2 border-foreground uppercase font-bold text-muted-foreground text-[9px] tracking-wider">
                <th className="pb-2">API Key Identifier</th>
                <th className="pb-2 text-right">Total Requests</th>
                <th className="pb-2 text-right">Allowed</th>
                <th className="pb-2 text-right">Blocked</th>
                <th className="pb-2 text-right">Avg Latency</th>
                <th className="pb-2 text-right">Status Rate</th>
              </tr>
            </thead>
            <tbody>
              {stats.map((item) => (
                <tr key={item.keyId} className="border-b border-foreground/10 hover:bg-muted/5 font-mono">
                  <td className="py-2.5 font-bold">
                    {item.keyId === "public / anonymous" ? (
                      <span className="text-muted-foreground uppercase text-[10px] italic">Public Client / Direct HTTP</span>
                    ) : (
                      <code className="text-xs">{item.keyId}</code>
                    )}
                  </td>
                  <td className="py-2.5 text-right font-bold">{item.total}</td>
                  <td className="py-2.5 text-right text-green-500 font-bold">{item.allowed}</td>
                  <td className="py-2.5 text-right text-red-500 font-bold">{item.blocked}</td>
                  <td className="py-2.5 text-right font-bold">{item.avgLatency}ms</td>
                  <td className="py-2.5 text-right">
                    <span
                      className={`inline-block px-1.5 py-0.5 text-[9px] font-bold uppercase border border-foreground ${
                        item.successRate > 90
                          ? "bg-green-500/10 text-green-500"
                          : item.successRate > 50
                          ? "bg-yellow-500/10 text-yellow-600"
                          : "bg-red-500/10 text-red-500"
                      }`}
                    >
                      {item.successRate}% OK
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </Panel>
  )
}
