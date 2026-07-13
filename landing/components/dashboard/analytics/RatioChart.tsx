"use client"

import * as React from "react"
import { ShieldAlert } from "lucide-react"
import {
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from "recharts"
import { Panel, PanelHeader, Label } from "@/components/dashboard/kit"

interface RatioChartProps {
  allowed: number
  blocked: number
}

export function RatioChart({ allowed, blocked }: RatioChartProps) {
  const total = allowed + blocked
  const blockedRate = total > 0 ? ((blocked / total) * 100).toFixed(1) : "0.0"

  const pieData = [
    { name: "ALLOWED", value: allowed, color: "#22c55e" },
    { name: "BLOCKED", value: blocked, color: "#ef4444" },
  ]

  return (
    <Panel>
      <PanelHeader title="Allowed vs Blocked" icon={ShieldAlert} />
      <div className="p-4 flex flex-col items-center justify-center h-[300px] relative">
        {total === 0 ? (
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
                <span>ALLOWED ({allowed})</span>
              </div>
              <div className="flex items-center gap-1.5">
                <div className="w-2.5 h-2.5 bg-red-500 border border-black" />
                <span>BLOCKED ({blocked})</span>
              </div>
            </div>
          </>
        )}
      </div>
    </Panel>
  )
}
