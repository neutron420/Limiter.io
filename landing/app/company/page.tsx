"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Navbar } from "@/components/navbar"
import { Footer } from "@/components/footer"
import { Compass, Network, Award, Heart } from "lucide-react"

const ease = [0.22, 1, 0.36, 1] as const

const PHILOSOPHIES = [
  {
    title: "Developer Autonomy",
    desc: "We believe system infrastructure must remain fully observable and controllable by developers. Open source core modules are key to developer confidence.",
    icon: Compass
  },
  {
    title: "Zero Ingress Friction",
    desc: "API rate limiting shouldn't dictate routing speed. We build algorithms co-located with memory states to guarantee sub-millisecond overhead.",
    icon: Network
  },
  {
    title: "Fail-Open Resiliency",
    desc: "Infrastructure guarantees should include clean fault isolation. If a gateway goes down, APIs must fail open to protect client traffic availability.",
    icon: Award
  },
  {
    title: "Open Metrics Integrity",
    desc: "API execution data belongs to the business. We deliver streaming analytics directly to client Kafka brokers without storing proprietary logs.",
    icon: Heart
  }
]

export default function CompanyPage() {
  return (
    <div className="min-h-screen dot-grid-bg">
      <Navbar />

      <main className="w-full px-6 py-12 lg:px-12">
        {/* Page Header */}
        <div className="mb-12">
          <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
            {"// COMPANY / TEAM_MANIFESTO"}
          </span>
          <h1 className="mt-2 text-3xl lg:text-5xl font-mono font-bold uppercase tracking-tight">
            INFRASTRUCTURE <span className="text-[#ea580c]">BUILT</span> FOR DEV SREs
          </h1>
          <p className="mt-4 max-w-2xl text-xs lg:text-sm font-mono text-muted-foreground leading-relaxed">
            Limiter.io was founded to solve a simple challenge: rate limiting that is fast, resilient, 
            and transparent. We are an open-source first engineering team.
          </p>
        </div>

        {/* Philosophy Blocks */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-12">
          {PHILOSOPHIES.map((p, i) => (
            <motion.div
              key={p.title}
              initial={{ opacity: 0, scale: 0.98 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.05, duration: 0.5, ease }}
              className="border-2 border-foreground p-6 bg-background flex flex-col justify-between"
            >
              <div className="flex items-start gap-4">
                <div className="border border-foreground/20 p-2 bg-[#ea580c]/5 text-[#ea580c]">
                  <p.icon size={16} />
                </div>
                <div className="flex-1">
                  <h3 className="text-sm font-mono font-bold uppercase tracking-wider">{p.title}</h3>
                  <p className="mt-2 text-xs font-mono text-muted-foreground leading-relaxed">{p.desc}</p>
                </div>
              </div>
            </motion.div>
          ))}
        </div>

        {/* Timeline block */}
        <div className="border-2 border-foreground p-6 bg-background/50 backdrop-blur-sm">
          <div className="flex items-center justify-between border-b-2 border-foreground pb-4 mb-6">
            <span className="text-xs font-mono font-bold uppercase">Our Milestone Roadmap</span>
            <span className="text-[10px] tracking-[0.2em] uppercase text-muted-foreground font-mono">
              PROJECT_EVOLUTION
            </span>
          </div>

          <div className="flex flex-col gap-6 font-mono text-xs">
            <div className="border-l-2 border-[#ea580c] pl-4 py-1">
              <span className="text-[#ea580c] font-bold uppercase">Q1 2026 — Distributed Core</span>
              <p className="text-muted-foreground mt-1 leading-relaxed">
                Initialized GORM schemas, built the Gin HTTP API, and completed the first five Lua-backed algorithms in Redis.
              </p>
            </div>
            <div className="border-l-2 border-[#ea580c] pl-4 py-1">
              <span className="text-[#ea580c] font-bold uppercase">Q2 2026 — Resiliency Upgrades</span>
              <p className="text-muted-foreground mt-1 leading-relaxed">
                Developed the fail-open SDK. Provisioned Kafka telemetry metrics pipelines and real-time dashboard updates over WebSockets.
              </p>
            </div>
            <div className="border-l-2 border-foreground/20 pl-4 py-1">
              <span className="text-muted-foreground uppercase">Q3 2026 — Edge Synchronization</span>
              <p className="text-muted-foreground mt-1 leading-relaxed">
                Replicate rate limits globally using geodistributed databases and Cloudflare Worker routing interceptors.
              </p>
            </div>
          </div>
        </div>
      </main>

      <Footer />
    </div>
  )
}
