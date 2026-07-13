"use client"

import * as React from "react"
import { Panel, PanelHeader } from "@/components/dashboard/kit"
import { Clock } from "lucide-react"

interface LogItem {
  id: string;
  latency_ms: number;
}

interface LatencyPercentilesProps {
  logs: LogItem[]
}

export function LatencyPercentiles({ logs }: LatencyPercentilesProps) {
  const percentiles = React.useMemo(() => {
    if (logs.length === 0) return { p50: 0, p90: 0, p99: 0 }
    const latencies = logs.map((l) => l.latency_ms).sort((a, b) => a - b)
    const getPercentile = (p: number) => {
      const idx = Math.min(latencies.length - 1, Math.floor((p / 100) * latencies.length))
      return latencies[idx]
    }
    return {
      p50: getPercentile(50),
      p90: getPercentile(90),
      p99: getPercentile(99),
    }
  }, [logs])

  return (
    <Panel>
      <PanelHeader title="Latency Percentiles" icon={Clock} />
      <div className="p-4 flex flex-col gap-4 min-h-[300px] justify-center">
        {logs.length === 0 ? (
          <div className="flex h-full flex-1 items-center justify-center text-muted-foreground uppercase text-[10px] my-auto">
            No latency telemetry captured yet.
          </div>
        ) : (
          <div className="grid grid-cols-3 gap-3">
            {[
              { label: "P50 (Median)", value: percentiles.p50, color: "bg-[#ea580c]/10 text-[#ea580c]" },
              { label: "P90 (Slowest 10%)", value: percentiles.p90, color: "bg-yellow-500/10 text-yellow-600" },
              { label: "P99 (Slowest 1%)", value: percentiles.p99, color: "bg-red-500/10 text-red-600" },
            ].map((p) => (
              <div key={p.label} className="border-2 border-foreground p-3 flex flex-col items-center justify-center text-center">
                <span className="text-[8px] uppercase tracking-wider text-muted-foreground font-bold leading-tight h-8 flex items-center justify-center">
                  {p.label}
                </span>
                <span className="text-xl font-bold font-mono tracking-tight mt-2 tabular-nums">
                  {p.value} ms
                </span>
                <div className={`mt-2 text-[8px] font-bold uppercase px-1.5 py-0.5 border border-foreground ${p.color}`}>
                  Verified
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </Panel>
  )
}
