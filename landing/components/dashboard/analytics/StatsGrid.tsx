"use client"

import * as React from "react"
import { Shield, ShieldAlert, Sparkles, Clock } from "lucide-react"
import { StatCard } from "@/components/dashboard/kit"

interface StatsGridProps {
  total: number
  allowed: number
  blocked: number
  avgLatency: number
}

export function StatsGrid({ total, allowed, blocked, avgLatency }: StatsGridProps) {
  return (
    <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        label="Total Requests"
        value={total.toLocaleString()}
        icon={Sparkles}
        hint="All incoming requests processed"
      />
      <StatCard
        label="Allowed Requests"
        value={allowed.toLocaleString()}
        icon={Shield}
        hint="Requests within quota rules"
      />
      <StatCard
        label="Blocked Requests"
        value={blocked.toLocaleString()}
        icon={ShieldAlert}
        hint="Requests blocked by policy"
      />
      <StatCard
        label="Avg Response Latency"
        value={`${avgLatency.toFixed(2)} ms`}
        icon={Clock}
        hint="Telemetry validation latency"
      />
    </div>
  )
}
