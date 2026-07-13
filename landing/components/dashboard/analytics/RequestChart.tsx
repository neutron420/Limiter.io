"use client"

import * as React from "react"
import { BarChart3 } from "lucide-react"
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
} from "recharts"
import { Panel, PanelHeader } from "@/components/dashboard/kit"

interface TimeSeriesItem {
  time: string
  allowed: number
  blocked: number
}

interface RequestChartProps {
  timeseries: TimeSeriesItem[]
  bucket: string
}

export function RequestChart({ timeseries, bucket }: RequestChartProps) {
  return (
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
  )
}
