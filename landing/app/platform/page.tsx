"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Navbar } from "@/components/navbar"
import { Footer } from "@/components/footer"
import { Cpu, Zap, Shield, HardDrive, Share2, Layers } from "lucide-react"

const ease = [0.22, 1, 0.36, 1] as const

const ALGORITHMS = [
  {
    name: "Token Bucket",
    desc: "Maintains a rolling counter of available tokens refilled at a constant rate. Supports instantaneous bursts without throttling.",
    useCase: "Standard API endpoints, user login routes, payment gates.",
    code: `local key = KEYS[1]
local limit = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])`
  },
  {
    name: "Fixed Window",
    desc: "Divides time into static windows (e.g. 1 minute) and tracks absolute request counts within that period. Discards window count upon boundary overlap.",
    useCase: "Daily scraping limits, monthly data sync boundaries.",
    code: `local count = redis.call("INCR", key)
if count == 1 then
  redis.call("EXPIRE", key, window)
end`
  },
  {
    name: "Sliding Window Counter",
    desc: "Uses a weighted average of the current and previous windows to compute the current rate, smoothing out boundary-crossing spikes.",
    useCase: "High-traffic web hooks, global ingress protection.",
    code: `local prev_count = redis.call("GET", prev_key) or 0
local curr_count = redis.call("GET", curr_key) or 0
local weight = (window_sec - elapsed) / window_sec`
  },
  {
    name: "Sliding Window Log",
    desc: "Logs every individual request timestamp in a Redis sorted set (ZSET). Evicts timestamps older than the window, offering complete accuracy.",
    useCase: "High-value financial transfers, sensitive auth validation.",
    code: `redis.call("ZREMRANGEBYSCORE", key, 0, min_score)
local current_requests = redis.call("ZCARD", key)
redis.call("ZADD", key, now, request_id)`
  },
  {
    name: "Leaky Bucket",
    desc: "Queues requests in a buffer that drips at a constant rate, smoothing out bursty traffic and enforcing a steady, strict output flow.",
    useCase: "External third-party API sync, batch process ingestion.",
    code: `local last_update = redis.call("HGET", key, "last")
local water = redis.call("HGET", key, "water")
local leaked = (now - last_update) * drip_rate`
  }
]

export default function PlatformPage() {
  return (
    <div className="min-h-screen dot-grid-bg">
      <Navbar />
      
      <main className="w-full px-6 py-12 lg:px-12">
        {/* Page Header */}
        <div className="mb-12">
          <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
            {"// PLATFORM / INFRASTRUCTURE_SPEC"}
          </span>
          <h1 className="mt-2 text-3xl lg:text-5xl font-mono font-bold uppercase tracking-tight">
            THE ENGINE BEHIND <span className="text-[#ea580c]">SUB-MS</span> QUOTAS
          </h1>
          <p className="mt-4 max-w-2xl text-xs lg:text-sm font-mono text-muted-foreground leading-relaxed">
            Limiter.io sits at your edge network and executes rate evaluation globally in less than a millisecond.
            By utilizing co-located memory segments and precompiled Redis Lua scripts, we guarantee complete concurrency safety.
          </p>
        </div>

        {/* Technical Architecture Block */}
        <div className="border-2 border-foreground p-6 mb-12 bg-background/50 backdrop-blur-sm">
          <div className="flex items-center justify-between border-b-2 border-foreground pb-4 mb-6">
            <span className="text-xs font-mono font-bold uppercase flex items-center gap-2">
              <Cpu size={14} className="text-[#ea580c]" />
              01. Atomic Execution Pipeline
            </span>
            <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
              SPEC_V4.9.0
            </span>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="flex flex-col gap-2">
              <span className="text-xs font-mono font-bold uppercase text-[#ea580c]">Multi-Tenant Isolation</span>
              <p className="text-xs font-mono text-muted-foreground leading-relaxed">
                Tokens and quotas are partitioned using cryptographically isolated namespaces. Tenant metrics never bleed across boundaries, guaranteeing safety.
              </p>
            </div>
            <div className="flex flex-col gap-2">
              <span className="text-xs font-mono font-bold uppercase text-[#ea580c]">Precompiled Lua Evaluation</span>
              <p className="text-xs font-mono text-muted-foreground leading-relaxed">
                Algorithm scripts are pre-loaded in memory using SHA hashes. Evaluated directly in Redis to prevent state discrepancies and race conditions.
              </p>
            </div>
            <div className="flex flex-col gap-2">
              <span className="text-xs font-mono font-bold uppercase text-[#ea580c]">Edge Proxy Fallback</span>
              <p className="text-xs font-mono text-muted-foreground leading-relaxed">
                Failures are bypassed gracefully. The client SDK features automatic fail-open strategies, maintaining API availability if backend clusters undergo updates.
              </p>
            </div>
          </div>
        </div>

        {/* Algorithms Directory */}
        <div className="mb-12">
          <div className="flex items-center gap-4 mb-8">
            <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
              {"// 02. ALGORITHM_STANDARDS"}
            </span>
            <div className="flex-1 border-t border-border" />
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {ALGORITHMS.map((algo, index) => (
              <motion.div
                key={algo.name}
                initial={{ opacity: 0, y: 16 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: index * 0.05, duration: 0.5, ease }}
                className="border-2 border-foreground flex flex-col justify-between"
              >
                <div className="border-b-2 border-foreground p-4 bg-muted/10">
                  <div className="flex items-center justify-between">
                    <h3 className="text-sm font-mono font-bold uppercase tracking-wider">{algo.name}</h3>
                    <span className="text-[9px] font-mono bg-foreground/5 text-muted-foreground px-2 py-0.5 border border-foreground/10 uppercase">
                      ACTIVE
                    </span>
                  </div>
                </div>

                <div className="p-4 flex-1 flex flex-col justify-between gap-4">
                  <div className="flex flex-col gap-2">
                    <p className="text-xs font-mono text-muted-foreground leading-relaxed">{algo.desc}</p>
                    <p className="text-[11px] font-mono text-foreground leading-relaxed">
                      <strong className="text-[#ea580c]">RECOMMENDED FOR:</strong> {algo.useCase}
                    </p>
                  </div>

                  <div className="bg-foreground text-background font-mono p-3 text-[10px] overflow-x-auto select-all border border-foreground/10">
                    <pre><code>{algo.code}</code></pre>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* Global Cluster Stats */}
        <div>
          <div className="flex items-center gap-4 mb-8">
            <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
              {"// 03. EDGE_LATENCY_METRICS"}
            </span>
            <div className="flex-1 border-t border-border" />
          </div>

          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            {[
              { title: "Avg Resolution", value: "0.24ms", icon: Zap },
              { title: "Cache Hit Rate", value: "99.98%", icon: Shield },
              { title: "Cluster Capacity", value: "10M rps", icon: Layers },
              { title: "Data Storage", value: "SSD NVMe", icon: HardDrive }
            ].map((stat, i) => (
              <div key={stat.title} className="border-2 border-foreground p-4 bg-background">
                <stat.icon size={16} className="text-[#ea580c] mb-2" />
                <span className="text-[9px] font-mono tracking-widest text-muted-foreground uppercase">{stat.title}</span>
                <div className="text-lg font-mono font-bold mt-1 uppercase">{stat.value}</div>
              </div>
            ))}
          </div>
        </div>
      </main>

      <Footer />
    </div>
  )
}
